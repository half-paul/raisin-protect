'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import {
  Grid3X3,
  Search,
  ChevronLeft,
  ChevronRight,
  CheckCircle2,
  Circle,
  Minus,
} from 'lucide-react';
import {
  MappingMatrixData,
  MappingMatrixFramework,
  MappingMatrixControl,
  getMappingMatrix,
} from '@/lib/api';

const STRENGTH_COLORS: Record<string, string> = {
  primary: 'bg-green-500 dark:bg-green-600',
  supporting: 'bg-blue-500 dark:bg-blue-600',
  partial: 'bg-amber-500 dark:bg-amber-600',
};

const STRENGTH_BG: Record<string, string> = {
  primary: 'bg-green-500/20 hover:bg-green-500/30',
  supporting: 'bg-blue-500/20 hover:bg-blue-500/30',
  partial: 'bg-amber-500/20 hover:bg-amber-500/30',
};

export default function MappingMatrixPage() {
  const [data, setData] = useState<MappingMatrixData | null>(null);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('');
  const perPage = 20;

  const fetchMatrix = useCallback(async () => {
    setLoading(true);
    try {
      const params: Record<string, string> = {
        page: String(page),
        per_page: String(perPage),
      };
      if (search) params.search = search;
      if (categoryFilter && categoryFilter !== 'all') params.control_category = categoryFilter;

      const res = await getMappingMatrix(params);
      setData(res.data);
      setTotal(res.meta?.total_controls || res.meta?.total || 0);
    } catch {
      // handle
    } finally {
      setLoading(false);
    }
  }, [page, search, categoryFilter]);

  useEffect(() => {
    fetchMatrix();
  }, [fetchMatrix]);

  // Debounced search
  useEffect(() => {
    const timer = setTimeout(() => {
      setSearch(searchInput);
      setPage(1);
    }, 300);
    return () => clearTimeout(timer);
  }, [searchInput]);

  const totalPages = Math.ceil(total / perPage);
  const frameworks = data?.frameworks || [];
  const controls = data?.controls || [];

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
          <Grid3X3 className="h-8 w-8" />
          Control Mapping Matrix
        </h1>
        <p className="text-muted-foreground mt-1">
          Cross-framework view — which controls satisfy which requirements across all your frameworks
        </p>
      </div>

      {/* Legend */}
      <div className="flex items-center gap-6 text-sm">
        <span className="text-muted-foreground">Mapping strength:</span>
        <div className="flex items-center gap-1.5">
          <div className="w-3 h-3 rounded-sm bg-green-500" />
          <span>Primary</span>
        </div>
        <div className="flex items-center gap-1.5">
          <div className="w-3 h-3 rounded-sm bg-blue-500" />
          <span>Supporting</span>
        </div>
        <div className="flex items-center gap-1.5">
          <div className="w-3 h-3 rounded-sm bg-amber-500" />
          <span>Partial</span>
        </div>
        <div className="flex items-center gap-1.5">
          <div className="w-3 h-3 rounded-sm bg-muted border" />
          <span>No mapping</span>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap items-center gap-4">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search controls..."
            className="pl-10"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
          />
        </div>
        <Select value={categoryFilter} onValueChange={(v) => { setCategoryFilter(v); setPage(1); }}>
          <SelectTrigger className="w-[160px]">
            <SelectValue placeholder="Category" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Categories</SelectItem>
            <SelectItem value="technical">Technical</SelectItem>
            <SelectItem value="administrative">Administrative</SelectItem>
            <SelectItem value="physical">Physical</SelectItem>
            <SelectItem value="operational">Operational</SelectItem>
          </SelectContent>
        </Select>
        <span className="text-sm text-muted-foreground">{total} controls</span>
      </div>

      {/* Matrix */}
      {loading ? (
        <div className="flex items-center justify-center py-24">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
        </div>
      ) : frameworks.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12 text-muted-foreground">
            <Grid3X3 className="h-12 w-12 mb-4 opacity-50" />
            <p className="text-lg font-medium">No activated frameworks</p>
            <p className="text-sm">Activate frameworks first to see the mapping matrix</p>
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardContent className="p-0 overflow-x-auto">
            <TooltipProvider delayDuration={200}>
              <table className="w-full border-collapse text-sm">
                <thead>
                  <tr className="border-b">
                    <th className="text-left p-3 font-medium sticky left-0 bg-card z-10 min-w-[200px]">
                      Control
                    </th>
                    {frameworks.map((fw) => (
                      <th
                        key={fw.id}
                        className="p-3 text-center font-medium min-w-[120px]"
                      >
                        <div className="text-xs">{fw.name}</div>
                        <div className="text-[10px] text-muted-foreground font-normal">
                          v{fw.version}
                        </div>
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {controls.map((ctrl) => (
                    <tr key={ctrl.id} className="border-b hover:bg-muted/50">
                      <td className="p-3 sticky left-0 bg-card z-10">
                        <Link
                          href={`/controls/${ctrl.id}`}
                          className="font-mono text-xs font-medium hover:text-primary"
                        >
                          {ctrl.identifier}
                        </Link>
                        <div className="text-xs text-muted-foreground line-clamp-1 mt-0.5">
                          {ctrl.title}
                        </div>
                      </td>
                      {frameworks.map((fw) => {
                        const mappings = ctrl.mappings_by_framework?.[fw.identifier] || [];
                        if (mappings.length === 0) {
                          return (
                            <td key={fw.id} className="p-3 text-center">
                              <div className="flex items-center justify-center">
                                <Minus className="h-3 w-3 text-muted-foreground/30" />
                              </div>
                            </td>
                          );
                        }

                        // Determine strongest mapping
                        const strengths = mappings.map((m) => m.strength);
                        const strongest = strengths.includes('primary')
                          ? 'primary'
                          : strengths.includes('supporting')
                          ? 'supporting'
                          : 'partial';

                        return (
                          <td key={fw.id} className="p-3 text-center">
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <div
                                  className={`inline-flex items-center justify-center w-8 h-8 rounded-md cursor-default ${STRENGTH_BG[strongest]}`}
                                >
                                  <span className="text-xs font-bold">{mappings.length}</span>
                                </div>
                              </TooltipTrigger>
                              <TooltipContent side="top" className="max-w-xs">
                                <div className="space-y-1">
                                  <p className="font-medium text-xs">
                                    {ctrl.identifier} → {fw.name}
                                  </p>
                                  {mappings.map((m, i) => (
                                    <div key={i} className="text-xs flex items-center gap-1.5">
                                      <div className={`w-2 h-2 rounded-full ${STRENGTH_COLORS[m.strength]}`} />
                                      <span className="font-mono">{m.identifier}</span>
                                      <span className="text-muted-foreground">({m.strength})</span>
                                    </div>
                                  ))}
                                </div>
                              </TooltipContent>
                            </Tooltip>
                          </td>
                        );
                      })}
                    </tr>
                  ))}
                  {controls.length === 0 && (
                    <tr>
                      <td colSpan={frameworks.length + 1} className="text-center py-12 text-muted-foreground">
                        No controls found
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </TooltipProvider>
          </CardContent>
        </Card>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <span className="text-sm text-muted-foreground">
            Page {page} of {totalPages}
          </span>
          <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      )}
    </div>
  );
}
