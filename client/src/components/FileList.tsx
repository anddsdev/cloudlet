import React from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import type { FileInfo, DirectoryInfo } from "@/services/api";
import {
  File,
  Folder,
  Download,
  Trash2,
  Edit,
  MoreHorizontal,
} from "lucide-react";
import { formatFileSize, formatDate } from "@/lib/formatters";

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
    ...directories.map((dir) => ({ ...dir, type: "directory" as const })),
    ...files.map((file) => ({ ...file, type: "file" as const })),
  ];

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[50px]"></TableHead>
            <TableHead>Name</TableHead>
            <TableHead>Size</TableHead>
            <TableHead>Items</TableHead>
            <TableHead>Modified</TableHead>
            <TableHead className="w-[100px]">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {allItems.length === 0 ? (
            <TableRow>
              <TableCell
                colSpan={6}
                className="text-center py-8 text-muted-foreground"
              >
                No files or folders found
              </TableCell>
            </TableRow>
          ) : (
            allItems.map((item) => (
              <TableRow
                key={item.path}
                className="hover:bg-muted/50 cursor-pointer"
                onClick={() =>
                  item.type === "directory"
                    ? onDirectoryClick(item.path)
                    : undefined
                }
              >
                <TableCell>
                  {item.type === "directory" ? (
                    <Folder className="h-5 w-5 text-blue-500" />
                  ) : (
                    <File className="h-5 w-5 text-gray-500" />
                  )}
                </TableCell>
                <TableCell className="font-medium">{item.name}</TableCell>
                <TableCell>
                  {item.type === "directory" &&
                  (item as DirectoryInfo).total_size > 0
                    ? formatFileSize((item as DirectoryInfo).total_size)
                    : formatFileSize((item as FileInfo).size)}
                </TableCell>
                <TableCell>
                  {item.type === "directory"
                    ? (item as DirectoryInfo).item_count
                    : "-"}
                </TableCell>
                <TableCell>
                  {item.type === "file"
                    ? formatDate((item as FileInfo).updated_at)
                    : formatDate((item as DirectoryInfo).updated_at)}
                </TableCell>
                <TableCell>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button
                        variant="ghost"
                        className="h-8 w-8 p-0"
                        onClick={(e) => e.stopPropagation()}
                      >
                        <MoreHorizontal className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem
                        onClick={(e) => {
                          e.stopPropagation();
                          onFileRename(item.path, item.name);
                        }}
                      >
                        <Edit className="mr-2 h-4 w-4" />
                        Edit
                      </DropdownMenuItem>
                      {item.type === "file" && (
                        <DropdownMenuItem
                          onClick={(e) => {
                            e.stopPropagation();
                            onFileDownload(item.path);
                          }}
                        >
                          <Download className="mr-2 h-4 w-4" />
                          Download
                        </DropdownMenuItem>
                      )}
                      <DropdownMenuItem
                        className="text-destructive focus:text-destructive"
                        onClick={(e) => {
                          e.stopPropagation();
                          onFileDelete(
                            item.path,
                            item.name,
                            item.type === "directory"
                          );
                        }}
                      >
                        <Trash2 className="mr-2 h-4 w-4" />
                        Delete
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
};
