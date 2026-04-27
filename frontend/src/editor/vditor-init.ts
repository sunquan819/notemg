import Vditor from 'vditor';

export function createVditor(
  containerId: string,
  options: {
    value?: string;
    onInput?: (value: string) => void;
    onUpload?: (files: File[]) => Promise<string | null>;
  }
): Vditor {
  const token = localStorage.getItem('access_token') || '';

  return new Vditor(containerId, {
    height: '100%',
    mode: 'ir',
    theme: 'dark',
    icon: 'ant',
    placeholder: 'Start writing...',
    value: options.value || '',
    cache: { enable: false },
    toolbar: [
      'headings', 'bold', 'italic', 'strike', '|',
      'list', 'ordered-list', 'check', 'outdent', 'indent', '|',
      'quote', 'code', 'inline-code', '|',
      'link', 'upload', 'table', '|',
      'undo', 'redo', '|',
      'fullscreen', 'edit-mode',
    ],
    upload: {
      url: '/api/attachments/upload',
      headers: { Authorization: `Bearer ${token}` },
      accept: 'image/*',
      handler: options.onUpload || (async () => null),
    },
    input: options.onInput,
  });
}
