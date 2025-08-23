import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { FileUploadZone } from './FileUploadZone';
import { FileList } from './FileList';
import { Breadcrumb } from './Breadcrumb';
import { CreateDirectoryDialog } from './CreateDirectoryDialog';
import { RenameDialog } from './RenameDialog';
import { DeleteConfirmDialog } from './DeleteConfirmDialog';
import { fileService } from '@/services/fileService';
import type { ListFilesResponse } from '@/services/api';
import { FolderPlus, RefreshCw, HardDrive } from 'lucide-react';
import { toast } from 'sonner';
import { formatFileSize } from '@/lib/formatters';

export const Dashboard: React.FC = () => {
  const [currentPath, setCurrentPath] = useState('/');
  const [data, setData] = useState<ListFilesResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [createDirOpen, setCreateDirOpen] = useState(false);
  const [renameOpen, setRenameOpen] = useState(false);
  const [renameTarget, setRenameTarget] = useState({ path: '', name: '' });
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState({ path: '', name: '', isDirectory: false, hasChildren: false });
  const [isDeleting, setIsDeleting] = useState(false);

  const loadFiles = async (path: string = currentPath) => {
    setIsLoading(true);
    try {
      const result = await fileService.listFiles(path);
      setData(result);
      setCurrentPath(path);
    } catch (error) {
      toast.error('Failed to load files');
      console.error('Error loading files:', error);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadFiles('/');
  }, []);

  const handleNavigate = (path: string) => {
    loadFiles(path);
  };

  const handleDirectoryClick = (path: string) => {
    loadFiles(path);
  };

  const handleFileDownload = (path: string) => {
    fileService.downloadFile(path);
  };

  const handleFileDelete = async (path: string, name: string, isDirectory: boolean) => {
    let hasChildren = false;
    
    // Check if directory has children
    if (isDirectory) {
      try {
        const dirInfo = await fileService.listFiles(path);
        hasChildren = (dirInfo.files?.length || 0) > 0 || (dirInfo.directories?.length || 0) > 0;
      } catch (error) {
        console.error('Error checking directory contents:', error);
      }
    }
    
    setDeleteTarget({ path, name, isDirectory, hasChildren });
    setDeleteOpen(true);
  };

  const handleConfirmDelete = async (recursive: boolean = false) => {
    setIsDeleting(true);
    try {
      await fileService.deleteFile(deleteTarget.path, recursive);
      toast.success('Item deleted successfully');
      setDeleteOpen(false);
      loadFiles();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to delete item';
      toast.error(errorMessage);
    } finally {
      setIsDeleting(false);
    }
  };

  const handleFileRename = (path: string, currentName: string) => {
    setRenameTarget({ path, name: currentName });
    setRenameOpen(true);
  };

  const handleCreateDirectory = async (name: string, parentPath: string) => {
    await fileService.createDirectory(name, parentPath);
  };

  const handleRename = async (path: string, newName: string) => {
    await fileService.renameFile(path, newName);
  };

  const refresh = () => {
    loadFiles();
  };

  if (isLoading && !data) {
    return (
      <div className="flex items-center justify-center h-96">
        <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8 max-w-6xl">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">☁️ Cloudlet</h1>
        <p className="text-muted-foreground">
          Simple, fast, and secure file storage
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-6">
          <Card>
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <CardTitle className="text-lg">Files & Folders</CardTitle>
                <div className="flex items-center space-x-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setCreateDirOpen(true)}
                  >
                    <FolderPlus className="h-4 w-4 mr-2" />
                    New Folder
                  </Button>
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={refresh}
                    disabled={isLoading}
                  >
                    <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
                  </Button>
                </div>
              </div>
              
              {data && data.breadcrumbs && (
                <Breadcrumb
                  items={data.breadcrumbs}
                  onNavigate={handleNavigate}
                />
              )}
            </CardHeader>
            
            <CardContent>
              {data && (
                <FileList
                  files={data.files || []}
                  directories={data.directories || []}
                  onDirectoryClick={handleDirectoryClick}
                  onFileDownload={handleFileDownload}
                  onFileDelete={handleFileDelete}
                  onFileRename={handleFileRename}
                />
              )}
            </CardContent>
          </Card>
        </div>

        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Upload Files</CardTitle>
            </CardHeader>
            <CardContent>
              <FileUploadZone
                path={currentPath}
                onUploadComplete={() => loadFiles()}
              />
            </CardContent>
          </Card>

          {data && (
            <Card>
              <CardHeader>
                <CardTitle className="text-lg flex items-center">
                  <HardDrive className="h-5 w-5 mr-2" />
                  Storage Info
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Total Files:</span>
                  <span className="font-medium">{data.total_files}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Total Folders:</span>
                  <span className="font-medium">{data.total_directories}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Total Size:</span>
                  <span className="font-medium">{formatFileSize(data.total_size)}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Current Path:</span>
                  <span className="font-medium font-mono text-xs">
                    {currentPath === '/' ? 'Root' : currentPath}
                  </span>
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      </div>

      <CreateDirectoryDialog
        open={createDirOpen}
        onOpenChange={setCreateDirOpen}
        currentPath={currentPath}
        onDirectoryCreated={() => loadFiles()}
        onCreateDirectory={handleCreateDirectory}
      />

      <RenameDialog
        open={renameOpen}
        onOpenChange={setRenameOpen}
        currentName={renameTarget.name}
        filePath={renameTarget.path}
        onRename={handleRename}
        onRenamed={() => loadFiles()}
      />

      <DeleteConfirmDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        itemName={deleteTarget.name}
        itemPath={deleteTarget.path}
        isDirectory={deleteTarget.isDirectory}
        onConfirm={handleConfirmDelete}
        isDeleting={isDeleting}
        hasChildren={deleteTarget.hasChildren}
      />
    </div>
  );
};