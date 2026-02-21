'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import { Label } from '@/components/ui/label';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { ArrowLeft, AlertTriangle, TrendingUp } from 'lucide-react';
import {
  HeatMapData, RiskSeverity, LikelihoodLevel, ImpactLevel,
  getRiskHeatMap,
} from '@/lib/api';
import {
  LIKELIHOOD_LABELS, LIKELIHOOD_ORDER,
  IMPACT_LABELS, IMPACT_ORDER,
  SEVERITY_LABELS, SEVERITY_COLORS,
  RISK_CATEGORY_LABELS,
  heatMapCellColor, scoreToSeverity,
} from '@/components/risk/constants';

export default function RiskHeatMapPage() {
  const [heatMap, setHeatMap] = useState<HeatMapData | null>(null);
  const [loading, setLoading] = useState(true);
  const [scoreType, setScoreType] = useState('residual');
  const [categoryFilter, setCategoryFilter] = useState('');

  const fetchHeatMap = useCallback(async () => {
    try {
      setLoading(true);
      const params: Record<string, string> = { score_type: scoreType };
      if (categoryFilter) params.category = categoryFilter;
      const res = await getRiskHeatMap(params);
      setHeatMap(res.data);
    } catch (err) {
      console.error('Failed to fetch heat map:', err);
    } finally {
      setLoading(false);
    }
  }, [scoreType, categoryFilter]);

  useEffect(() => { fetchHeatMap(); }, [fetchHeatMap]);

  // Build lookup
  const lookup = new Map<string, HeatMapData['grid'][0]>();
  if (heatMap) {
    for (const cell of heatMap.grid) {
      lookup.set(`${cell.likelihood_score}-${cell.impact_score}`, cell);
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Link href="/risks">
          <Button variant="ghost" size="icon"><ArrowLeft className="h-4 w-4" /></Button>
        </Link>
        <div className="flex-1">
          <h1 className="text-2xl font-bold">Risk Heat Map</h1>
          <p className="text-sm text-muted-foreground">5×5 likelihood vs impact grid visualization</p>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap items-end gap-4">
        <div className="w-[160px]">
          <Label className="text-xs">Score Type</Label>
          <Select value={scoreType} onValueChange={setScoreType}>
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="residual">Residual</SelectItem>
              <SelectItem value="inherent">Inherent</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="w-[180px]">
          <Label className="text-xs">Category</Label>
          <Select value={categoryFilter} onValueChange={v => setCategoryFilter(v === 'all' ? '' : v)}>
            <SelectTrigger><SelectValue placeholder="All" /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All categories</SelectItem>
              {Object.entries(RISK_CATEGORY_LABELS).map(([k, v]) => (
                <SelectItem key={k} value={k}>{v}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      {loading ? (
        <div className="flex items-center justify-center h-64 text-muted-foreground">Loading heat map...</div>
      ) : heatMap ? (
        <>
          {/* Summary Cards */}
          <div className="grid gap-4 md:grid-cols-5">
            <Card>
              <CardContent className="p-4">
                <div className="text-2xl font-bold">{heatMap.summary.total_risks}</div>
                <p className="text-xs text-muted-foreground">Total Risks</p>
              </CardContent>
            </Card>
            {(['critical', 'high', 'medium', 'low'] as RiskSeverity[]).map(sev => (
              <Card key={sev}>
                <CardContent className="p-4">
                  <div className="flex items-center gap-2">
                    <div className="text-2xl font-bold">{heatMap.summary.by_severity[sev] || 0}</div>
                    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SEVERITY_COLORS[sev]}`}>
                      {SEVERITY_LABELS[sev]}
                    </span>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          {/* Heat Map Grid */}
          <Card>
            <CardHeader>
              <CardTitle className="capitalize">{scoreType} Risk Heat Map</CardTitle>
              <CardDescription>
                {heatMap.summary.total_risks} risks · Average score: {heatMap.summary.average_score?.toFixed(1)} · {heatMap.summary.appetite_breaches} appetite breaches
              </CardDescription>
            </CardHeader>
            <CardContent>
              <TooltipProvider>
                <div className="grid grid-cols-[120px_repeat(5,1fr)] gap-2 max-w-2xl mx-auto">
                  {/* Column headers */}
                  <div className="flex items-end justify-center pb-2">
                    <span className="text-xs text-muted-foreground font-medium -rotate-0">Likelihood ↑</span>
                  </div>
                  {IMPACT_ORDER.map(imp => (
                    <div key={imp} className="text-center text-xs font-medium text-muted-foreground pb-2">
                      {IMPACT_LABELS[imp]}
                    </div>
                  ))}

                  {/* Grid rows — almost_certain at top */}
                  {[...LIKELIHOOD_ORDER].reverse().map(lik => {
                    const lScore = LIKELIHOOD_ORDER.indexOf(lik) + 1;
                    return (
                      <>
                        <div key={`label-${lik}`} className="flex items-center text-xs font-medium text-muted-foreground pr-2 justify-end">
                          {LIKELIHOOD_LABELS[lik]}
                        </div>
                        {IMPACT_ORDER.map(imp => {
                          const iScore = IMPACT_ORDER.indexOf(imp) + 1;
                          const cellScore = lScore * iScore;
                          const cellKey = `${lScore}-${iScore}`;
                          const cellData = lookup.get(cellKey);
                          const count = cellData?.count || 0;
                          return (
                            <Tooltip key={`${lik}-${imp}`}>
                              <TooltipTrigger asChild>
                                <div
                                  className={`h-16 rounded-md flex flex-col items-center justify-center cursor-default transition-all hover:ring-2 hover:ring-primary/50 ${heatMapCellColor(cellScore)}`}
                                >
                                  <span className={`text-lg font-bold ${count > 0 ? 'text-white dark:text-white' : 'text-muted-foreground/40'}`}>
                                    {count > 0 ? count : ''}
                                  </span>
                                  <span className="text-[10px] text-muted-foreground/60">{cellScore}</span>
                                </div>
                              </TooltipTrigger>
                              <TooltipContent className="max-w-xs">
                                <p className="font-semibold">
                                  {LIKELIHOOD_LABELS[lik]} × {IMPACT_LABELS[imp]} = {cellScore}
                                  <span className={`ml-2 inline-flex items-center rounded-full px-1.5 py-0.5 text-[10px] font-medium ${SEVERITY_COLORS[scoreToSeverity(cellScore)]}`}>
                                    {SEVERITY_LABELS[scoreToSeverity(cellScore)]}
                                  </span>
                                </p>
                                <p className="text-xs">{count} risk{count !== 1 ? 's' : ''} in this cell</p>
                                {cellData?.risks?.map(r => (
                                  <p key={r.id} className="text-xs mt-0.5">
                                    <span className="font-mono">{r.identifier}</span> — {r.title}
                                  </p>
                                ))}
                              </TooltipContent>
                            </Tooltip>
                          );
                        })}
                      </>
                    );
                  })}

                  {/* Impact label */}
                  <div />
                  <div className="col-span-5 text-center text-xs font-medium text-muted-foreground pt-2">
                    Impact →
                  </div>
                </div>
              </TooltipProvider>
            </CardContent>
          </Card>

          {/* Legend */}
          <Card>
            <CardContent className="p-4">
              <h4 className="text-sm font-semibold mb-2">Severity Legend</h4>
              <div className="flex flex-wrap gap-4 text-sm">
                <div className="flex items-center gap-2">
                  <div className="w-6 h-4 rounded bg-red-500/80" />
                  <span>Critical (20-25)</span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-6 h-4 rounded bg-orange-400/70" />
                  <span>High (12-19)</span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-6 h-4 rounded bg-yellow-300/60" />
                  <span>Medium (6-11)</span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-6 h-4 rounded bg-green-200/50" />
                  <span>Low (1-5)</span>
                </div>
              </div>
            </CardContent>
          </Card>
        </>
      ) : (
        <div className="text-center text-muted-foreground py-12">No heat map data available</div>
      )}
    </div>
  );
}
