import type { UploadResponse, BatchUploadResponse, BatchProgressResponse } from './api';

const API_BASE_URL = '/api/v1';

export class UploadService {
  async uploadSingle(file: File, path: string = '/'): Promise<UploadResponse> {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('path', path);

    const response = await fetch(`${API_BASE_URL}/upload`, {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      throw new Error(`Upload failed: ${response.status}`);
    }

    return response.json();
  }

  async uploadMultiple(files: File[], path: string = '/'): Promise<UploadResponse> {
    const formData = new FormData();
    files.forEach(file => {
      formData.append('files', file);
    });
    formData.append('path', path);

    const response = await fetch(`${API_BASE_URL}/upload/multiple`, {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      throw new Error(`Multiple upload failed: ${response.status}`);
    }

    return response.json();
  }

  async uploadStream(file: File, path: string = '/'): Promise<UploadResponse> {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('path', path);

    const response = await fetch(`${API_BASE_URL}/upload/stream`, {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      throw new Error(`Stream upload failed: ${response.status}`);
    }

    return response.json();
  }

  async uploadBatch(files: File[], path: string = '/'): Promise<BatchUploadResponse> {
    const formData = new FormData();
    files.forEach(file => {
      formData.append('files', file);
    });
    formData.append('path', path);

    const response = await fetch(`${API_BASE_URL}/upload/batch`, {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      throw new Error(`Batch upload failed: ${response.status}`);
    }

    return response.json();
  }

  async getBatchProgress(batchId: string): Promise<BatchProgressResponse> {
    const response = await fetch(`${API_BASE_URL}/upload/batch/${batchId}/progress`);

    if (!response.ok) {
      throw new Error(`Failed to get batch progress: ${response.status}`);
    }

    return response.json();
  }

  async cancelBatch(batchId: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/upload/batch/${batchId}`, {
      method: 'DELETE',
    });

    if (!response.ok) {
      throw new Error(`Failed to cancel batch: ${response.status}`);
    }
  }
}