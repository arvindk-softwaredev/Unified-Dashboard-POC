export const queryKeys = {
  repositories: () => ["repositories"] as const,
  trackingTrends: () => ["tracking-trends"] as const,
  repoInsights: (owner: string, name: string, refreshTrigger: number, aiMode: boolean) =>
    ["repo-insights", owner, name, refreshTrigger, aiMode] as const,
};
