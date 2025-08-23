import { UploadService } from './uploadService';
import type { UploadResponse, BatchUploadResponse } from './api';

export interface UploadStrategy {
  upload(files: File[], path: string): Promise<UploadResponse | BatchUploadResponse>;
}

export class SingleUploadStrategy implements UploadStrategy {
  private uploadService: UploadService;

  constructor() {
    this.uploadService = new UploadService();
  }

  async upload(files: File[], path: string): Promise<UploadResponse> {
    if (files.length === 0) throw new Error('No files to upload');
    return this.uploadService.uploadSingle(files[0], path);
  }
}

export class MultipleUploadStrategy implements UploadStrategy {
  private uploadService: UploadService;

  constructor() {
    this.uploadService = new UploadService();
  }

  async upload(files: File[], path: string): Promise<UploadResponse> {
    return this.uploadService.uploadMultiple(files, path);
  }
}

export class StreamUploadStrategy implements UploadStrategy {
  private uploadService: UploadService;

  constructor() {
    this.uploadService = new UploadService();
  }

  async upload(files: File[], path: string): Promise<UploadResponse> {
    if (files.length === 0) throw new Error('No files to upload');
    return this.uploadService.uploadStream(files[0], path);
  }
}

export class BatchUploadStrategy implements UploadStrategy {
  private uploadService: UploadService;

  constructor() {
    this.uploadService = new UploadService();
  }

  async upload(files: File[], path: string): Promise<BatchUploadResponse> {
    return this.uploadService.uploadBatch(files, path);
  }
}

export class UploadStrategyContext {
  private strategy: UploadStrategy;
  
  private static readonly LARGE_FILE_SIZE = 10 * 1024 * 1024; // 10MB
  private static readonly MANY_FILES_THRESHOLD = 10;
  private static readonly TOTAL_SIZE_THRESHOLD = 500 * 1024 * 1024; // 500MB

  constructor(strategy?: UploadStrategy) {
    this.strategy = strategy || new SingleUploadStrategy();
  }

  setStrategy(strategy: UploadStrategy): void {
    this.strategy = strategy;
  }

  determineStrategy(files: File[]): UploadStrategy {
    if (files.length === 0) {
      return new SingleUploadStrategy();
    }

    const totalSize = files.reduce((sum, file) => sum + file.size, 0);
    const hasLargeFiles = files.some(file => file.size > UploadStrategyContext.LARGE_FILE_SIZE);

    if (files.length === 1) {
      if (hasLargeFiles) {
        return new StreamUploadStrategy();
      }
      return new SingleUploadStrategy();
    }

    if (files.length >= UploadStrategyContext.MANY_FILES_THRESHOLD || 
        totalSize > UploadStrategyContext.TOTAL_SIZE_THRESHOLD) {
      return new BatchUploadStrategy();
    }

    if (hasLargeFiles) {
      return new BatchUploadStrategy();
    }

    return new MultipleUploadStrategy();
  }

  async uploadFiles(files: File[], path: string, autoDetect: boolean = true): Promise<UploadResponse | BatchUploadResponse> {
    if (autoDetect) {
      this.strategy = this.determineStrategy(files);
    }

    return this.strategy.upload(files, path);
  }
}