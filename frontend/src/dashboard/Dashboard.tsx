import Alert from "@mui/material/Alert";
import Box from "@mui/material/Box";
import CircularProgress from "@mui/material/CircularProgress";
import { useEffect, useState } from "react";
import { fetchRepositories, fetchTrackingTrends, type Repository, type TrendPoint } from "../api/client";
import { buildTrendsFromRepos, mergeTrendPoints } from "./buildTrendsFromRepos";
import { useAiMode } from "../theme/AiModeContext";
import { DashboardLayout } from "./DashboardLayout";
import { DashboardStats } from "./DashboardStats";
import { Header } from "./Header";
import { RepositoriesTable } from "./RepositoriesTable";

export function Dashboard() {
  const { aiMode } = useAiMode();
  const [repos, setRepos] = useState<Repository[]>([]);
  const [org, setOrg] = useState("tektoncd");
  const [loading, setLoading] = useState(true);
  const [reposError, setReposError] = useState<string | null>(null);
  const [trendPoints, setTrendPoints] = useState<TrendPoint[]>([]);
  const [trendsLoading, setTrendsLoading] = useState(true);
  const [trendsWarning, setTrendsWarning] = useState<string | null>(null);

  useEffect(() => {
    fetchRepositories({ aiMode })
      .then(data => {
        setRepos(data.repositories);
        setOrg(data.organization);
      })
      .catch(err => setReposError(err instanceof Error ? err.message : "Failed to load repositories"))
      .finally(() => setLoading(false));
  }, []);

  return (
    <DashboardLayout>
      <Box
        sx={{
          display: "flex",
          flexDirection: "column",
          height: { xs: "calc(100vh - 120px)", md: "calc(100vh - 140px)" },
          minHeight: 400,
        }}
      >
        <Box sx={{ flexShrink: 0, pb: 2 }}>
          <Header
            title="Repositories"
            organization={org}
            subtitle="Browse tektoncd projects, then expand a row for workload charts and tracked issues."
          />
          {reposError && <Alert severity="error">{reposError}</Alert>}
        </Box>

        {loading ? (
          <Box sx={{ display: "flex", justifyContent: "center", flex: 1, alignItems: "center" }}>
            <CircularProgress />
          </Box>
        ) : (
          <>
            <DashboardStats
              repositories={repos}
              organization={org}
              trendPoints={trendPoints}
              trendsLoading={trendsLoading}
              trendsWarning={trendsWarning}
            />
            <RepositoriesTable repositories={repos} />
          </>
        )}
      </Box>
    </DashboardLayout>
  );
}
