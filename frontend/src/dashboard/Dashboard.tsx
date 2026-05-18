import Alert from "@mui/material/Alert";
import Box from "@mui/material/Box";
import CircularProgress from "@mui/material/CircularProgress";
import { useEffect, useState } from "react";
import { fetchRepositories, type Repository } from "../api/client";
import { DashboardLayout } from "./DashboardLayout";
import { Header } from "./Header";
import { RepositoriesTable } from "./RepositoriesTable";

export function Dashboard() {
  const [repos, setRepos] = useState<Repository[]>([]);
  const [org, setOrg] = useState("tektoncd");
  const [loading, setLoading] = useState(true);
  const [reposError, setReposError] = useState<string | null>(null);

  useEffect(() => {
    fetchRepositories()
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
            subtitle={`${org} — expand a row for donut, monthly trends, and item list`}
          />
          {reposError && <Alert severity="error">{reposError}</Alert>}
        </Box>

        {loading ? (
          <Box sx={{ display: "flex", justifyContent: "center", flex: 1, alignItems: "center" }}>
            <CircularProgress />
          </Box>
        ) : (
          <RepositoriesTable repositories={repos} />
        )}
      </Box>
    </DashboardLayout>
  );
}
