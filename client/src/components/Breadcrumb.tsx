import React from 'react';
import { Button } from '@/components/ui/button';
import { ChevronRight, Home } from 'lucide-react';

interface BreadcrumbItem {
  name: string;
  path: string;
}

interface BreadcrumbProps {
  items: BreadcrumbItem[];
  onNavigate: (path: string) => void;
}

export const Breadcrumb: React.FC<BreadcrumbProps> = ({ items, onNavigate }) => {
  return (
    <nav className="flex items-center space-x-1 text-sm text-muted-foreground mb-4">
      <Button
        variant="ghost"
        size="sm"
        className="h-8 px-2"
        onClick={() => onNavigate('/')}
      >
        <Home className="h-4 w-4" />
      </Button>
      
      {items.map((item) => (
        <React.Fragment key={item.path}>
          <ChevronRight className="h-4 w-4" />
          <Button
            variant="ghost"
            size="sm"
            className="h-8 px-2"
            onClick={() => onNavigate(item.path)}
          >
            {item.name}
          </Button>
        </React.Fragment>
      ))}
    </nav>
  );
};