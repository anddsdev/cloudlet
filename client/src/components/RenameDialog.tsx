import React, { useState, useEffect } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { toast } from 'sonner';

interface RenameDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  currentName: string;
  filePath: string;
  onRename: (path: string, newName: string) => Promise<void>;
  onRenamed: () => void;
}

export const RenameDialog: React.FC<RenameDialogProps> = ({
  open,
  onOpenChange,
  currentName,
  filePath,
  onRename,
  onRenamed,
}) => {
  const [newName, setNewName] = useState('');
  const [isRenaming, setIsRenaming] = useState(false);

  useEffect(() => {
    if (open) {
      setNewName(currentName);
    }
  }, [open, currentName]);

  const handleRename = async () => {
    if (!newName.trim()) {
      toast.error('Please enter a name');
      return;
    }

    if (newName.trim() === currentName) {
      onOpenChange(false);
      return;
    }

    setIsRenaming(true);
    try {
      await onRename(filePath, newName.trim());
      toast.success(`Successfully renamed to "${newName}"`);
      onOpenChange(false);
      onRenamed();
    } catch (error) {
      toast.error('Failed to rename');
    } finally {
      setIsRenaming(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Rename</DialogTitle>
          <DialogDescription>
            Enter a new name for "{currentName}"
          </DialogDescription>
        </DialogHeader>
        
        <div className="py-4">
          <Input
            placeholder="New name"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                handleRename();
              }
            }}
            autoFocus
          />
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isRenaming}
          >
            Cancel
          </Button>
          <Button
            onClick={handleRename}
            disabled={isRenaming || !newName.trim()}
          >
            {isRenaming ? 'Renaming...' : 'Rename'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};