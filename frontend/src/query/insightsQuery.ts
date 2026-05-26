import { useQuery } from "@tanstack/react-query";
import { fetchRepoInsights } from "../api/client";
import type { CategoryInsight } from "../api/client";
import { queryKeys } from "./keys";

function defaultCategoryKey(list: CategoryInsight[]): string {
  return list.find(c => c.total > 0)?.key ?? list[0]?.key ?? "good_first_issues";
}

export function useRepoInsightsQuery(
  owner: string,
  name: string,
  refreshTrigger: number,
  aiMode: boolean,
  enabled: boolean,
) {
  return useQuery({
    queryKey: queryKeys.repoInsights(owner, name, refreshTrigger, aiMode),
    queryFn: async () => {
      const refresh = refreshTrigger > 0;
      const result = await fetchRepoInsights(owner, name, { refresh, aiMode });
      const list = result.data.categories ?? [];
      return {
        categories: list,
        latestRelease: result.data.latest_release ?? null,
        selectedKey: defaultCategoryKey(list),
        fromCache: result.fromCache,
      };
    },
    enabled: enabled && Boolean(owner) && Boolean(name),
  });
}
