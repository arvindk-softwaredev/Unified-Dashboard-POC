import Box from "@mui/material/Box";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import CardHeader from "@mui/material/CardHeader";
import Grid from "@mui/material/Grid";
import Typography from "@mui/material/Typography";
import { BarChart } from "@mui/x-charts/BarChart";
import { PieChart } from "@mui/x-charts/PieChart";
import type { RepoMetricPoint, TrackingMetrics } from "../api/client";

type OverviewChartsProps = {
  summary: TrackingMetrics;
  breakdown: RepoMetricPoint[];
};

function StatCard({ label, value, suffix = "" }: { label: string; value: number | string; suffix?: string }) {
  return (
    <Card variant="outlined" sx={{ height: "100%" }}>
      <CardContent>
        <Typography variant="overline" color="text.secondary">
          {label}
        </Typography>
        <Typography variant="h4" sx={{ fontWeight: 600 }}>
          {value}
          {suffix}
        </Typography>
      </CardContent>
    </Card>
  );
}

export function OverviewCharts({ summary, breakdown }: OverviewChartsProps) {
  const pieData = [
    { id: 0, label: "Good first issues", value: summary.good_first_issues, color: "#1976d2" },
    { id: 1, label: "Open bugs", value: summary.open_bugs, color: "#d32f2f" },
    { id: 2, label: "Pending PRs", value: summary.pending_prs, color: "#ed6c02" },
  ].filter(d => d.value > 0);

  const repoNames = breakdown.map(b => b.repository);
  const ciFailure = breakdown.map(b => b.metrics.ci_failure_rate);
  const avgCI = breakdown.map(b => b.metrics.avg_ci_minutes);

  return (
    <Box>
      <Grid container spacing={2}>
        <Grid size={{ xs: 6, sm: 4, md: 2 }}>
          <StatCard label="Good first issues" value={summary.good_first_issues} />
        </Grid>
        <Grid size={{ xs: 6, sm: 4, md: 2 }}>
          <StatCard label="Open bugs" value={summary.open_bugs} />
        </Grid>
        <Grid size={{ xs: 6, sm: 4, md: 2 }}>
          <StatCard label="Pending PRs" value={summary.pending_prs} />
        </Grid>
        <Grid size={{ xs: 6, sm: 4, md: 2 }}>
          <StatCard label="Linked bugs" value={summary.linked_bugs} />
        </Grid>
        <Grid size={{ xs: 6, sm: 4, md: 2 }}>
          <StatCard label="Dependency alerts" value={summary.dependency_alerts} />
        </Grid>
        <Grid size={{ xs: 6, sm: 6, md: 3 }}>
          <StatCard label="CI failure rate" value={summary.ci_failure_rate.toFixed(1)} suffix="%" />
        </Grid>
        <Grid size={{ xs: 6, sm: 6, md: 3 }}>
          <StatCard label="Avg CI time" value={summary.avg_ci_minutes.toFixed(1)} suffix=" min" />
        </Grid>
      </Grid>

      <Grid container spacing={2} sx={{ mt: 1 }}>
        <Grid size={{ xs: 12, md: 4 }}>
          <Card variant="outlined">
            <CardHeader title="Upstream workload" subheader="Org-wide issue & PR mix" />
            <CardContent>
              {pieData.length > 0 ? (
                <PieChart
                  series={[{ data: pieData, innerRadius: 40, outerRadius: 100, paddingAngle: 2, cornerRadius: 4 }]}
                  height={220}
                  margin={{ top: 0, bottom: 0, left: 0, right: 0 }}
                />
              ) : (
                <Typography variant="body2" color="text.secondary">
                  No tracked items yet.
                </Typography>
              )}
            </CardContent>
          </Card>
        </Grid>
        <Grid size={{ xs: 12, md: 8 }}>
          <Card variant="outlined">
            <CardHeader title="CI failure rate by repository" subheader="Last workflow runs sampled per repo" />
            <CardContent>
              {breakdown.length > 0 ? (
                <BarChart
                  xAxis={[{ scaleType: "band", data: repoNames }]}
                  series={[{ data: ciFailure, label: "Failure %", color: "#d32f2f" }]}
                  height={240}
                  margin={{ left: 48, right: 16, top: 16, bottom: 56 }}
                />
              ) : (
                <Typography variant="body2" color="text.secondary">
                  Per-repo CI metrics load when you open a repository.
                </Typography>
              )}
            </CardContent>
          </Card>
        </Grid>
        <Grid size={{ xs: 12 }}>
          <Card variant="outlined">
            <CardHeader title="Average CI execution time" subheader="Minutes per workflow run (top repos)" />
            <CardContent>
              {breakdown.length > 0 ? (
                <BarChart
                  xAxis={[{ scaleType: "band", data: repoNames }]}
                  series={[{ data: avgCI, label: "Minutes", color: "#2e7d32" }]}
                  height={260}
                  margin={{ left: 48, right: 16, top: 16, bottom: 56 }}
                />
              ) : (
                <Typography variant="body2" color="text.secondary">
                  No CI timing data available.
                </Typography>
              )}
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
}
