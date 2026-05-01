declare module 'vditor' {
  export default class Vditor {
    constructor(id: string, options?: VditorOptions);
    getValue(): string;
    setValue(value: string, clearStack?: boolean): void;
    insertValue(value: string): void;
    destroy(): void;
    getHTML(): string;
  }

  interface VditorOptions {
    height?: string | number;
    mode?: string;
    theme?: string;
    icon?: string;
    placeholder?: string;
    value?: string;
    cache?: { enable?: boolean };
    toolbar?: (string | number)[];
    options?: {
      markdown?: {
        toc?: boolean;
        mark?: boolean;
        footnotes?: boolean;
        autoSpace?: boolean;
      };
      math?: {
        inlineDigit?: boolean;
        engine?: string;
      };
    };
    upload?: {
      url?: string;
      headers?: Record<string, string>;
      accept?: string;
      handler?: (files: File[]) => any;
    };
    input?: (value: string) => void;
    after?: () => void;
  }
}

declare module 'vditor/dist/index.css';
