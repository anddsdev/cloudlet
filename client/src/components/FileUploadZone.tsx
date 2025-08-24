import React, { useCallback, useState } from 'react';
import { useDropzone } from 'react-dropzone';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Progress } from '@/components/ui/progress';
import { Upload, X, FileText, AlertCircle } from 'lucide-react';
import { UploadStrategyContext } from '@/services/uploadStrategy';
import { toast } from 'sonner';

interface FileUploadZoneProps {
  path: string;
  onUploadComplete: () => void;
}

interface UploadingFile {
  file: File;
  progress: number;
  status: 'uploading' | 'completed' | 'error';
  error?: string;
}

export const FileUploadZone: React.FC<FileUploadZoneProps> = ({
  path,
  onUploadComplete,
}) => {
  const [uploadingFiles, setUploadingFiles] = useState<UploadingFile[]>([]);
  const [isUploading, setIsUploading] = useState(false);

  const uploadContext = new UploadStrategyContext();

  const onDrop = useCallback(async (acceptedFiles: File[]) => {
    if (acceptedFiles.length === 0) return;

    setIsUploading(true);
    const initialFiles: UploadingFile[] = acceptedFiles.map(file => ({
      file,
      progress: 0,
      status: 'uploading' as const,
    }));
    
    setUploadingFiles(initialFiles);

    try {
      const strategy = uploadContext.determineStrategy(acceptedFiles);
      const strategyName = strategy.constructor.name;
      
      toast.info(`Using ${strategyName.replace('Strategy', '')} upload for ${acceptedFiles.length} file(s)`);

      await uploadContext.uploadFiles(acceptedFiles, path);

      setUploadingFiles(prev => 
        prev.map(item => ({
          ...item,
          progress: 100,
          status: 'completed' as const,
        }))
      );

      toast.success(`Successfully uploaded ${acceptedFiles.length} file(s)`);
      onUploadComplete();
      
      setTimeout(() => {
        setUploadingFiles([]);
      }, 2000);

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Upload failed';
      
      setUploadingFiles(prev => 
        prev.map(item => ({
          ...item,
          status: 'error' as const,
          error: errorMessage,
        }))
      );

      toast.error(`Upload failed: ${errorMessage}`);
    } finally {
      setIsUploading(false);
    }
  }, [path, onUploadComplete]);

  const {
    getRootProps,
    getInputProps,
    isDragActive,
    isDragReject,
  } = useDropzone({
    onDrop,
    disabled: isUploading,
  });

  const removeFile = (index: number) => {
    setUploadingFiles(prev => prev.filter((_, i) => i !== index));
  };

  return (
    <div className="space-y-4">
      <Card 
        className={`border-2 border-dashed transition-colors cursor-pointer ${
          isDragActive 
            ? 'border-primary bg-primary/5' 
            : isDragReject
            ? 'border-destructive bg-destructive/5'
            : 'border-muted-foreground/25 hover:border-muted-foreground/50'
        } ${isUploading ? 'opacity-50 cursor-not-allowed' : ''}`}
      >
        <CardContent 
          {...getRootProps()}
          className="flex flex-col items-center justify-center py-12 px-6 text-center"
        >
          <input {...getInputProps()} />
          <Upload className={`h-12 w-12 mb-4 ${
            isDragActive ? 'text-primary' : 'text-muted-foreground'
          }`} />
          
          {isDragActive ? (
            <div className="space-y-2">
              <p className="text-lg font-medium text-primary">Drop files here</p>
              <p className="text-sm text-muted-foreground">
                Release to upload files
              </p>
            </div>
          ) : (
            <div className="space-y-2">
              <p className="text-lg font-medium">Drag & drop files here</p>
              <p className="text-sm text-muted-foreground">
                or click to browse files
              </p>
              <Button variant="outline" size="sm" className="mt-4">
                Browse Files
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      {uploadingFiles.length > 0 && (
        <Card>
          <CardContent className="p-4">
            <h3 className="font-medium mb-3">Uploading Files</h3>
            <div className="space-y-3">
              {uploadingFiles.map((item, index) => (
                <div key={index} className="flex items-center space-x-3">
                  <FileText className="h-4 w-4 text-muted-foreground flex-shrink-0" />
                  
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between mb-1">
                      <p className="text-sm font-medium truncate">
                        {item.file.name}
                      </p>
                      <span className="text-xs text-muted-foreground">
                        {(item.file.size / 1024 / 1024).toFixed(1)} MB
                      </span>
                    </div>
                    
                    {item.status === 'uploading' && (
                      <Progress value={item.progress} className="h-2" />
                    )}
                    
                    {item.status === 'error' && (
                      <div className="flex items-center space-x-1 text-destructive">
                        <AlertCircle className="h-3 w-3" />
                        <span className="text-xs">{item.error}</span>
                      </div>
                    )}
                    
                    {item.status === 'completed' && (
                      <div className="text-xs text-green-600">
                        ✓ Upload complete
                      </div>
                    )}
                  </div>

                  {item.status !== 'uploading' && (
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-6 w-6"
                      onClick={() => removeFile(index)}
                    >
                      <X className="h-3 w-3" />
                    </Button>
                  )}
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
};