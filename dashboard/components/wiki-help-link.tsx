'use client';

import { HelpCircle } from 'lucide-react';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';

const WIKI_URL = process.env.NEXT_PUBLIC_WIKI_URL || 'http://localhost:3020';

interface WikiHelpLinkProps {
  path: string;
  label?: string;
}

export function WikiHelpLink({ path, label }: WikiHelpLinkProps) {
  const url = `${WIKI_URL}/${path}`;

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <a
            href={url}
            target="_blank"
            rel="noopener noreferrer"
            aria-label={label || 'View documentation'}
            className="inline-flex items-center justify-center h-7 w-7 rounded-md hover:bg-accent hover:text-accent-foreground transition-colors"
          >
            <HelpCircle className="h-4 w-4 text-muted-foreground" />
          </a>
        </TooltipTrigger>
        <TooltipContent>
          <p>{label || 'View documentation'}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
