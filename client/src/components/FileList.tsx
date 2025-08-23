import React from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import type { FileInfo, DirectoryInfo } from '@/services/api';
import {
  File,
  Folder,
  Download,
  Trash2,
  Edit,
} from 'lucide-react';
import { formatFileSize, formatDate } from '@/lib/formatters';

interface FileListProps {
  files: FileInfo[];
  directories: DirectoryInfo[];
  onDirectoryClick: (path: string) => void;
  onFileDownload: (path: string) => void;
  onFileDelete: (path: string, name: string, isDirectory: boolean) => void;
  onFileRename: (path: string, currentName: string) => void;
}

export const FileList: React.FC<FileListProps> = ({
  files,
  directories,
  onDirectoryClick,
  onFileDownload,
  onFileDelete,
  onFileRename,
}) => {
  const allItems = [
    ...directories.map(dir => ({ ...dir, type: 'directory' as const })),
    ...files.map(file => ({ ...file, type: 'file' as const }))
  ];

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[50px]"></TableHead>
            <TableHead>Name</TableHead>
            <TableHead>Size</TableHead>
            <TableHead>Modified</TableHead>
            <TableHead className="w-[100px]">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {allItems.length === 0 ? (
            <TableRow>
              <TableCell colSpan={5} className="text-center py-8 text-muted-foreground">
                No files or folders found
              </TableCell>
            </TableRow>
          ) : (
            allItems.map((item) => (
              <TableRow 
                key={item.path}
                className="hover:bg-muted/50 cursor-pointer"
                onClick={() => item.type === 'directory' ? onDirectoryClick(item.path) : undefined}
              >
                <TableCell>
                  {item.type === 'directory' ? (
                    <Folder className="h-5 w-5 text-blue-500" />
                  ) : (
                    <File className="h-5 w-5 text-gray-500" />
                  )}
                </TableCell>
                <TableCell className="font-medium">
                  {item.name}
                </TableCell>
                <TableCell>
                  {item.type === 'directory' 
                    ? `${(item as DirectoryInfo).file_count} items`
                    : formatFileSize((item as FileInfo).size)
                  }
                </TableCell>
                <TableCell>
                  {item.type === 'file' 
                    ? formatDate((item as FileInfo).modified)
                    : '-'
                  }
                </TableCell>
                <TableCell>
                  <div className="flex items-center space-x-1">
                    {item.type === 'file' && (
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8"
                        onClick={(e) => {
                          e.stopPropagation();
                          onFileDownload(item.path);
                        }}
                      >
                        <Download className="h-4 w-4" />
                      </Button>
                    )}
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8"
                      onClick={(e) => {
                        e.stopPropagation();
                        onFileRename(item.path, item.name);
                      }}
                    >
                      <Edit className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8 text-destructive hover:text-destructive"
                      onClick={(e) => {
                        e.stopPropagation();
                        onFileDelete(item.path, item.name, item.type === 'directory');
                      }}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
};