import { fetcher } from './api';
import type { ListFilesResponse } from './api';

export const fileService = {
  async listFiles(path: string = ''): Promise<ListFilesResponse> {
    const cleanPath = path === '/' ? '' : path;
    const url = cleanPath ? `/files/${encodeURIComponent(cleanPath)}` : '/files';
    return fetcher(url);
  },

  async createDirectory(name: string, parentPath: string = '/'): Promise<void> {
    return fetcher('/directories', {
      method: 'POST',
      body: JSON.stringify({ name, parent_path: parentPath }),
    });
  },

  async deleteFile(path: string, recursive: boolean = false): Promise<void> {
    const url = `/files${path.startsWith('/') ? path : '/' + path}${recursive ? '?recursive=true' : ''}`;
    return fetcher(url, {
      method: 'DELETE',
    });
  },

  async moveFile(sourcePath: string, destinationPath: string): Promise<void> {
    return fetcher('/move', {
      method: 'POST',
      body: JSON.stringify({ source_path: sourcePath, destination_path: destinationPath }),
    });
  },

  async renameFile(path: string, newName: string): Promise<void> {
    return fetcher('/rename', {
      method: 'POST',
      body: JSON.stringify({ path, new_name: newName }),
    });
  },

  async downloadFile(path: string): Promise<void> {
    const url = `/api/v1/download/${encodeURIComponent(path)}`;
    window.open(url, '_blank');
  },
};