const API_BASE_URL = "/api/v1";

export const fetcher = async (url: string, options?: RequestInit) => {
  const response = await fetch(`${API_BASE_URL}${url}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
  });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  return response.json();
};

export interface FileInfo {
  name: string;
  size: number;
  modified: string;
  is_directory: boolean;
  path: string;
  mime_type?: string;
}

export interface DirectoryInfo {
  name: string;
  path: string;
  file_count: number;
  total_size: number;
}

export interface ListFilesResponse {
  path: string;
  parent_path: string;
  files: FileInfo[];
  directories: DirectoryInfo[];
  total_files: number;
  total_directories: number;
  total_size: number;
  breadcrumbs: { name: string; path: string }[];
}

export interface UploadResponse {
  success: boolean;
  files: Array<{
    filename: string;
    path: string;
    size: number;
    error?: string;
  }>;
  errors?: string[];
}

export interface BatchUploadResponse {
  batch_id: string;
  status: "processing" | "completed" | "failed";
  total_files: number;
  processed_files: number;
  success_count: number;
  error_count: number;
  errors?: string[];
}

export interface BatchProgressResponse {
  batch_id: string;
  status: "processing" | "completed" | "failed" | "cancelled";
  total_files: number;
  processed_files: number;
  success_count: number;
  error_count: number;
  progress_percentage: number;
  estimated_time_remaining?: number;
  current_file?: string;
}
