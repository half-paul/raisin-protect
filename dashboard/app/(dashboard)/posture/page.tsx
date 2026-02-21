'use client';

import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Shield, TrendingUp, TrendingDown, Minus, CheckCircle2, XCircle,
  HelpCircle, RefreshCw,
} from 'lucide-react';
import { PostureData, getMonitoringPosture } from '@/lib/api';
import { cn } from '@/lib/utils';

function ScoreRing({ score, size = 120 }: { score: number; size?: number }) {
  const radius = (size - 12) / 2;
  const circumference = 2 * Math.PI * radius;
  const offset = circumference - (score / 100) * circumference;
  const color = score >= 80 ? 'text-green-500' : score >= 60 ? 'text-amber-500' : 'text-red-500';
  const strokeColor = score >= 80 ? 'stroke-green-500' : score >= 60 ? 'stroke-amber-500' : 'stroke-red-500';

  return (
    <div className="relative inline-flex items-center justify-center" style={{ width: size, height: size }}>
      <svg className="transform -rotate-90" width={size} height={size}>
        <circle
          cx={size / 2} cy={size / 2} r={radius}
          stroke="currentColor"
          strokeWidth="6"
          fill="transparent"
          className="text-muted/30"
        />
        <circle
          cx={size / 2} cy={size / 2} r={radius}
          strokeWidth="6"
          fill="transparent"
          strokeDasharray={circumference}
          strokeDashoffset={offset}
          strokeLinecap="round"
          className={cn('transition-all duration-700', strokeColor)}
        />
      </svg>
      <div className="absolute flex flex-col items-center">
        <span className={cn('text-2xl font-bold', color)}>{score.toFixed(1)}%</span>
      </div>
    </div>
  );
}

export default function PosturePage() {
  const [posture, setPosture] = useState<PostureData | null>(null);
  const [loading, setLoading] = useState(true);

  async function fetchData() {
    setLoading(true);
    try {
      const res = await getMonitoringPosture();
      setPosture(res.data);
    } catch {
      // handle
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { fetchData(); }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center py-24">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <Shield className="h-8 w-8" />
            Compliance Posture
          </h1>
          <p className="text-muted-foreground mt-1">
            Real-time compliance scores per activated framework
          </p>
        </div>
        <Button variant="outline" onClick={fetchData}>
          <RefreshCw className="h-4 w-4 mr-2" /> Refresh
        </Button>
      </div>

      {/* Overall score */}
      {posture && (
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-8">
              <ScoreRing score={posture.overall_score} size={140} />
              <div>
                <h2 className="text-xl font-semibold">Overall Posture Score</h2>
                <p className="text-muted-foreground text-sm mt-1">
                  Weighted average across {posture.frameworks.length} activated framework{posture.frameworks.length !== 1 ? 's' : ''}
                </p>
                <p className="text-sm mt-2">
                  {posture.overall_score >= 80
                    ? 'Strong compliance posture — maintain this level.'
                    : posture.overall_score >= 60
                    ? 'Moderate compliance posture — areas need attention.'
                    : 'Weak compliance posture — immediate action required.'}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Framework scores */}
      {posture && (
        <div className="grid gap-4 md:grid-cols-2">
          {posture.frameworks.map((fw) => (
            <Card key={fw.framework_id}>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle className="text-lg">{fw.framework_name}</CardTitle>
                    <CardDescription>Version {fw.framework_version}</CardDescription>
                  </div>
                  <ScoreRing score={fw.posture_score} size={80} />
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                {/* Progress bar */}
                <div className="space-y-1">
                  <div className="h-3 rounded-full bg-secondary overflow-hidden flex">
                    <div
                      className="h-full bg-green-500 transition-all"
                      style={{ width: `${(fw.passing / Math.max(fw.total_mapped_controls, 1)) * 100}%` }}
                    />
                    <div
                      className="h-full bg-red-500 transition-all"
                      style={{ width: `${(fw.failing / Math.max(fw.total_mapped_controls, 1)) * 100}%` }}
                    />
                    <div
                      className="h-full bg-gray-300 dark:bg-gray-600 transition-all"
                      style={{ width: `${(fw.untested / Math.max(fw.total_mapped_controls, 1)) * 100}%` }}
                    />
                  </div>
                </div>

                {/* Stats */}
                <div className="grid grid-cols-4 gap-2 text-center">
                  <div>
                    <p className="text-lg font-bold">{fw.total_mapped_controls}</p>
                    <p className="text-xs text-muted-foreground">Total</p>
                  </div>
                  <div>
                    <p className="text-lg font-bold text-green-600 dark:text-green-400">{fw.passing}</p>
                    <p className="text-xs text-muted-foreground">Passing</p>
                  </div>
                  <div>
                    <p className="text-lg font-bold text-red-600 dark:text-red-400">{fw.failing}</p>
                    <p className="text-xs text-muted-foreground">Failing</p>
                  </div>
                  <div>
                    <p className="text-lg font-bold text-gray-500">{fw.untested}</p>
                    <p className="text-xs text-muted-foreground">Untested</p>
                  </div>
                </div>

                {/* Trend */}
                {fw.trend && (
                  <div className="flex items-center gap-4 text-sm border-t pt-3">
                    <div className={cn(
                      'flex items-center gap-1',
                      fw.trend.direction === 'improving' ? 'text-green-600 dark:text-green-400' :
                      fw.trend.direction === 'declining' ? 'text-red-600 dark:text-red-400' :
                      'text-muted-foreground'
                    )}>
                      {fw.trend.direction === 'improving' && <TrendingUp className="h-4 w-4" />}
                      {fw.trend.direction === 'declining' && <TrendingDown className="h-4 w-4" />}
                      {fw.trend.direction === 'stable' && <Minus className="h-4 w-4" />}
                      <span className="capitalize">{fw.trend.direction}</span>
                    </div>
                    <span className="text-muted-foreground">
                      7d ago: {fw.trend['7d_ago'].toFixed(1)}%
                    </span>
                    <span className="text-muted-foreground">
                      30d ago: {fw.trend['30d_ago'].toFixed(1)}%
                    </span>
                  </div>
                )}
              </CardContent>
            </Card>
          ))}

          {posture.frameworks.length === 0 && (
            <div className="col-span-2 text-center py-12 text-muted-foreground">
              <Shield className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p className="text-lg font-medium">No frameworks activated</p>
              <p className="text-sm">Activate frameworks to see posture scores</p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
