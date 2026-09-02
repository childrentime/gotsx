// 由 gotsx 从 Go 宿主模块反射生成 —— 不要手改

declare module "host:data" {
  export const models: ModelStore;
}

declare module "host:intl" {
  export function fmtDate(arg0: string): string;
  export function fmtNumber(arg0: number): string;
  export function now(): string;
}

declare module "host:data" {
  export interface ModelStore {
    get(arg0: string): Model;
    like(arg0: string): number;
    list(): Model[];
    search(arg0: string): Model[];
  }
  export interface Model {
    id: string;
    title: string;
    author: string;
    desc: string;
    likes: number;
    tags: string[];
    createdAt: string;
  }
}
