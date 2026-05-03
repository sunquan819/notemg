package server

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/notemg/notemg/internal/config"
	"github.com/notemg/notemg/internal/handler"
	"github.com/notemg/notemg/internal/markdown"
	"github.com/notemg/notemg/internal/search"
	"github.com/notemg/notemg/internal/security"
	"github.com/notemg/notemg/internal/store"
)

type Server struct {
	cfg       *config.Config
	server    *http.Server
	auth      *security.Auth
	dbStore   *store.Store
	noteSt    *store.NoteStore
	folderSt  *store.FolderStore
	tagSt     *store.TagStore
	driveSt   *store.DriveStore
	md        *markdown.Engine
	searcher  *search.Searcher
	frontFS   embed.FS
}

func New(cfg *config.Config, frontFS embed.FS) *Server {
	return &Server{
		cfg:     cfg,
		frontFS: frontFS,
	}
}

func (s *Server) Start() error {
	if err := s.cfg.EnsureDataDirs(); err != nil {
		return err
	}

	s.auth = security.NewAuth(s.cfg)

	var err error
	s.dbStore, err = store.New(s.cfg)
	if err != nil {
		return err
	}

	s.noteSt = store.NewNoteStore(s.dbStore.DB(), s.cfg)
	s.folderSt = store.NewFolderStore(s.dbStore.DB())
	s.tagSt = store.NewTagStore(s.dbStore.DB())
	s.driveSt = store.NewDriveStore(s.dbStore.DB(), s.cfg)
	if err := s.driveSt.Init(); err != nil {
		log.Printf("Warning: drive init failed: %v", err)
	}
	s.md = markdown.NewEngine()

	s.searcher, err = search.NewSearcher(s.cfg.IndexPath())
	if err != nil {
		log.Printf("Warning: search index init failed: %v", err)
	}

	mux := s.setupRouter()

	s.server = &http.Server{
		Addr:         s.cfg.Addr(),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("NoteMG starting on %s", s.cfg.Addr())

	if s.cfg.Security.EnableHTTPS && s.cfg.Security.CertFile != "" {
		return s.server.ListenAndServeTLS(s.cfg.Security.CertFile, s.cfg.Security.KeyFile)
	}
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.searcher != nil {
		s.searcher.Close()
	}
	if s.dbStore != nil {
		s.dbStore.Close()
	}
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

func (s *Server) setupRouter() http.Handler {
	r := newRouter()

	r.Use(requestID)
	r.Use(realIP)
	r.Use(logger)
	r.Use(recoverer)
	r.Use(cors(s.cfg))

	api := r.Group("/api")

	authH := handler.NewAuthHandler(s.auth, s.cfg)
	api.Post("/auth/login", authH.Login)
	api.Post("/auth/init", authH.Init)
	api.Get("/auth/status", authH.Status)
	api.Post("/auth/refresh", authH.Refresh)
	authRequired := api.Group("")
	authRequired.Use(s.authMW())
	authRequired.Put("/auth/password", authH.ChangePassword)

	notes := api.Group("")
	notes.Use(s.authMW())

	noteH := handler.NewNoteHandler(s.noteSt, s.searcher, s.cfg)
	notes.Get("/notes", noteH.List)
	notes.Post("/notes", noteH.Create)
	notes.Get("/notes/{id}", noteH.Get)
	notes.Put("/notes/{id}", noteH.Update)
	notes.Delete("/notes/{id}", noteH.Delete)
	notes.Post("/notes/{id}/move", noteH.Move)
	notes.Post("/notes/{id}/duplicate", noteH.Duplicate)
	notes.Post("/notes/{id}/restore", noteH.Restore)
	notes.Delete("/notes/{id}/permanent", noteH.PermanentDelete)

	folderH := handler.NewFolderHandler(s.folderSt)
	notes.Get("/folders", folderH.Tree)
	notes.Post("/folders", folderH.Create)
	notes.Put("/folders/{id}", folderH.Update)
	notes.Delete("/folders/{id}", folderH.Delete)

	tagH := handler.NewTagHandler(s.tagSt)
	notes.Get("/tags", tagH.List)
	notes.Post("/tags", tagH.Create)
	notes.Delete("/tags/{id}", tagH.Delete)
	notes.Get("/tags/{id}/notes", tagH.NotesByTag)

	searchH := handler.NewSearchHandler(s.searcher)
	notes.Get("/search", searchH.Search)

	imageH := handler.NewImageHandler(s.cfg)
	notes.Post("/attachments/upload", imageH.Upload)
	notes.Get("/attachments/{id}", imageH.Get)
	notes.Delete("/attachments/{id}", imageH.Delete)

	mdH := handler.NewMarkdownHandler(s.md)
	notes.Post("/markdown/render", mdH.Render)

	ieH := handler.NewImportExportHandler(s.noteSt, s.md, s.cfg)
	notes.Post("/import/markdown", ieH.ImportMarkdown)
	notes.Post("/import/zip", ieH.ImportZip)
	notes.Get("/export/notes/{id}", ieH.ExportNote)
	notes.Post("/export/batch", ieH.ExportBatch)

	driveH := handler.NewDriveHandler(s.driveSt, s.cfg)
	notes.Get("/drive/files", driveH.List)
	notes.Post("/drive/folders", driveH.CreateFolder)
	notes.Post("/drive/upload", driveH.Upload)
	notes.Get("/drive/files/{id}", driveH.Get)
	notes.Get("/drive/files/{id}/preview", driveH.Preview)
	notes.Get("/drive/files/{id}/download", driveH.Download)
	notes.Get("/drive/files/{id}/thumbnail", driveH.Thumbnail)
	notes.Put("/drive/files/{id}", driveH.Rename)
	notes.Post("/drive/files/{id}/move", driveH.Move)
	notes.Delete("/drive/files/{id}", driveH.Delete)
	notes.Get("/drive/search", driveH.Search)

	spaFS, err := fs.Sub(s.frontFS, "frontend/dist")
	if err != nil {
		log.Printf("Warning: frontend dist not found in embed, API-only mode")
	} else {
		fileServer := http.FileServer(http.FS(spaFS))
		r.Get("/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			path := req.URL.Path

			if len(path) >= 4 && path[:4] == "/api" {
				http.NotFound(w, req)
				return
			}

			cleanPath := path
			if cleanPath == "/" {
				cleanPath = "/index.html"
			}

			if _, statErr := fs.Stat(spaFS, cleanPath[1:]); statErr != nil {
				req.URL.Path = "/index.html"
			}

			w.Header().Set("Cache-Control", "no-cache")
			fileServer.ServeHTTP(w, req)
		}))
	}

	return r
}

func (s *Server) authMW() Middleware {
	return security.AuthMiddleware(s.auth)
}
