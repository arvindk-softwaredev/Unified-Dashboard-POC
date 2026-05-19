import { useQuery } from "@tanstack/react-query";
import { useEffect, useRef } from "react";
import { fetchRepositories, fetchTrackingTrends } from "../api/client";
import { queryKeys } from "./keys";
import { queryClient } from "./queryClient";

/** Fetches org repositories — updates UI as soon as this query settles. */
export function useRepositoriesQuery() {
  return useQuery({
    queryKey: queryKeys.repositories(),
    queryFn: () => fetchRepositories(),
  });
}

/** Fetches monthly trends in parallel with repos (separate HTTP request). */
export function useTrackingTrendsQuery() {
  return useQuery({
    queryKey: queryKeys.trackingTrends(),
    queryFn: () => fetchTrackingTrends(),
  });
}

/** Re-run trends once repos are cached server-side (parallel first load may race). */
export function useRefetchTrackingTrendsAfterRepos(reposReady: boolean) {
  const didRefetch = useRef(false);

  useEffect(() => {
    if (!reposReady || didRefetch.current) return;
    didRefetch.current = true;
    void queryClient.invalidateQueries({ queryKey: queryKeys.trackingTrends() });
  }, [reposReady]);
}
