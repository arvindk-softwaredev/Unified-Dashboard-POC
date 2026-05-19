export const queryKeys = {
  repositories: () => ["repositories"] as const,
  trackingTrends: () => ["tracking-trends"] as const,
  repoInsights: (owner: string, name: string, refresh: boolean, aiMode: boolean) =>
    ["repo-insights", owner, name, refresh, aiMode] as const,
};
