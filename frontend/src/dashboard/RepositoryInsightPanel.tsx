import CachedIcon from "@mui/icons-material/Cached";
import OpenInNewIcon from "@mui/icons-material/OpenInNew";
import RefreshIcon from "@mui/icons-material/Refresh";
import Alert from "@mui/material/Alert";
import Box from "@mui/material/Box";
import Chip from "@mui/material/Chip";
import Fade from "@mui/material/Fade";
import Grid from "@mui/material/Grid";
import IconButton from "@mui/material/IconButton";
import Link from "@mui/material/Link";
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemText from "@mui/material/ListItemText";
import Stack from "@mui/material/Stack";
import Tooltip from "@mui/material/Tooltip";
import Typography from "@mui/material/Typography";
import { BarChart } from "@mui/x-charts/BarChart";
import { PieChart } from "@mui/x-charts/PieChart";
import { useCallback, useEffect, useMemo, useState } from "react";
import {
  fetchRepoInsights,
  type CategoryInsight,
  type Issue,
  type PullRequestItem,
  type Repository,
} from "../api/client";
import { InsightsLoading } from "./InsightsLoading";

const CHART_COLORS = ["#1976d2", "#d32f2f", "#ed6c02", "#2e7d32", "#9c27b0"];
const LOADING_DELAY_MS = 150;

/** Backend sends months as `YYYY-MM` (UTC). */
function formatMonthLabel(yearMonth: string): string {
  const match = /^(\d{4})-(\d{2})$/.exec(yearMonth);
  if (!match) return yearMonth;
  const year = Number(match[1]);
  const month = Number(match[2]);
  return new Date(year, month - 1, 1).toLocaleDateString("en-US", {
    month: "long",
    year: "numeric",
  });
}

function monthRangeLabel(months: string[]): string | null {
  if (months.length === 0) return null;
  const first = formatMonthLabel(months[0]!);
  const last = formatMonthLabel(months[months.length - 1]!);
  return first === last ? first : `${first} – ${last}`;
}

type RepositoryInsightPanelProps = {
  repo: Repository;
  /** Increment to force refresh from GitHub (bypass server cache). */
  refreshTrigger: number;
};

export function RepositoryInsightPanel({ repo, refreshTrigger }: RepositoryInsightPanelProps) {
  const [categories, setCategories] = useState<CategoryInsight[]>([]);
  const [selectedKey, setSelectedKey] = useState<string>("good_first_issues");
  const [ready, setReady] = useState(false);
  const [showLoader, setShowLoader] = useState(false);
  const [fromCache, setFromCache] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const applyInsights = useCallback((data: Awaited<ReturnType<typeof fetchRepoInsights>>["data"]) => {
    const list = Array.isArray(data.categories) ? data.categories : [];
    setCategories(list);
    const first = list.find(c => c.total > 0)?.key ?? list[0]?.key ?? "good_first_issues";
    setSelectedKey(first);
  }, []);

  useEffect(() => {
    const [owner, name] = repo.full_name.split("/");
    let cancelled = false;
    const forceRefresh = refreshTrigger > 0;

    setError(null);
    setFromCache(false);
    setReady(false);
    setCategories([]);
    setShowLoader(forceRefresh);

    let loadingTimer: ReturnType<typeof setTimeout> | undefined;
    if (!forceRefresh) {
      loadingTimer = setTimeout(() => {
        if (!cancelled) setShowLoader(true);
      }, LOADING_DELAY_MS);
    }

    fetchRepoInsights(owner, name, { refresh: forceRefresh })
      .then(({ data, fromCache: cached }) => {
        if (cancelled) return;
        clearTimeout(loadingTimer);
        applyInsights(data);
        setFromCache(cached);
        setShowLoader(false);
        setReady(true);
      })
      .catch(err => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "Failed to load insights");
          setShowLoader(false);
          setReady(true);
        }
      });

    return () => {
      cancelled = true;
      clearTimeout(loadingTimer);
    };
  }, [repo.full_name, refreshTrigger, applyInsights]);

  const safeCategories = categories ?? [];

  const selected = useMemo(
    () => safeCategories.find(c => c.key === selectedKey) ?? safeCategories[0],
    [safeCategories, selectedKey],
  );

  const pieData = useMemo(
    () =>
      safeCategories.map((c, i) => ({
        id: c.key,
        label: c.label,
        value: c.total,
        color: CHART_COLORS[i % CHART_COLORS.length],
      })),
    [safeCategories],
  );

  if (!ready) {
    if (showLoader) return <InsightsLoading />;
    return <Box sx={{ minHeight: 16 }} />;
  }

  if (error) {
    return <Alert severity="error">{error}</Alert>;
  }

  if (safeCategories.length === 0 || !selected) {
    return <Alert severity="info">No insight data for this repository.</Alert>;
  }

  const months = selected.monthly.map(m => m.month);
  const monthLabels = months.map(formatMonthLabel);
  const monthRange = monthRangeLabel(months);
  const counts = selected.monthly.map(m => m.count);

  return (
    <Fade in timeout={400}>
      <Box sx={{ p: 2, bgcolor: "action.hover", borderTop: 1, borderColor: "divider" }}>
        <Stack direction="row" sx={{ alignItems: "center", justifyContent: "space-between", mb: 1 }}>
          <Stack direction="row" spacing={1} sx={{ alignItems: "center" }}>
            <Typography variant="subtitle2">Upstream insights</Typography>
            {fromCache && refreshTrigger === 0 && (
              <Tooltip title="Loaded from server cache (up to 15 min old). Use Refresh for latest data.">
                <Chip icon={<CachedIcon />} label="Cached" size="small" variant="outlined" />
              </Tooltip>
            )}
          </Stack>
        </Stack>
        <Grid container spacing={2}>
          <Grid size={{ xs: 12, md: 4 }}>
            <Typography variant="subtitle2" gutterBottom>
              Workload mix — click a segment
            </Typography>
            <PieChart
              series={[
                {
                  data: pieData,
                  innerRadius: 50,
                  outerRadius: 90,
                  paddingAngle: 2,
                  cornerRadius: 3,
                  highlightScope: { fade: "global", highlight: "item" },
                  faded: { additionalRadius: -8 },
                },
              ]}
              height={220}
              onItemClick={(_e, item) => {
                if (item?.dataIndex != null && pieData[item.dataIndex]) {
                  setSelectedKey(String(pieData[item.dataIndex].id));
                }
              }}
            />
          </Grid>
          <Grid size={{ xs: 12, md: 8 }}>
            <Typography variant="subtitle2" gutterBottom>
              {selected.label} opened by month
            </Typography>
            {monthRange && (
              <Typography variant="caption" color="text.secondary" display="block" sx={{ mb: 0.5 }}>
                {monthRange} · by GitHub creation date (open items only)
              </Typography>
            )}
            {months.length > 0 ? (
              <BarChart
                xAxis={[{ scaleType: "band", data: monthLabels }]}
                series={[{ data: counts, label: selected.label, color: CHART_COLORS[0] }]}
                height={220}
                margin={{ left: 40, right: 12, top: 12, bottom: 48 }}
              />
            ) : (
              <Typography variant="body2" color="text.secondary">
                No dated items for this category.
              </Typography>
            )}
          </Grid>
          <Grid size={{ xs: 12 }}>
            <Typography variant="subtitle2" gutterBottom>
              {selected.label} ({selected.total})
            </Typography>
            <List dense disablePadding sx={{ maxHeight: 200, overflow: "auto" }}>
              <CategoryItemList category={selected} />
            </List>
          </Grid>
        </Grid>
      </Box>
    </Fade>
  );
}

function CategoryItemList({ category }: { category: CategoryInsight }) {
  if (category.key === "pending_prs") {
    if (!category.pull_requests?.length) {
      return <Typography variant="body2" color="text.secondary">No items.</Typography>;
    }
    return (
      <>
        {category.pull_requests.map((pr: PullRequestItem) => (
          <ListItem
            key={pr.id}
            divider
            secondaryAction={
              <IconButton
                component="a"
                href={pr.html_url}
                target="_blank"
                rel="noopener noreferrer"
                size="small"
                aria-label="Open on GitHub"
              >
                <OpenInNewIcon fontSize="small" />
              </IconButton>
            }
          >
            <ListItemText
              primary={
                <Link href={pr.html_url} target="_blank" rel="noopener noreferrer" underline="hover">
                  #{pr.number} {pr.title}
                </Link>
              }
            />
          </ListItem>
        ))}
      </>
    );
  }

  if (category.key === "dependency_alerts") {
    return (
      <Typography variant="body2" color="text.secondary">
        {category.total} open Dependabot alert(s). Enable GITHUB_TOKEN for alert details.
      </Typography>
    );
  }

  if (!category.issues?.length) {
    return <Typography variant="body2" color="text.secondary">No items.</Typography>;
  }

  return (
    <>
      {category.issues.map((issue: Issue) => (
        <ListItem
          key={issue.id}
          divider
          secondaryAction={
            <IconButton
              component="a"
              href={issue.html_url}
              target="_blank"
              rel="noopener noreferrer"
              size="small"
              aria-label="Open on GitHub"
            >
              <OpenInNewIcon fontSize="small" />
            </IconButton>
          }
        >
          <ListItemText
            primary={
              <Link href={issue.html_url} target="_blank" rel="noopener noreferrer" underline="hover">
                #{issue.number} {issue.title}
              </Link>
            }
          />
        </ListItem>
      ))}
    </>
  );
}
