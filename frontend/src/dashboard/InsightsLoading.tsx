import Box from "@mui/material/Box";
import Fade from "@mui/material/Fade";
import LinearProgress from "@mui/material/LinearProgress";
import Skeleton from "@mui/material/Skeleton";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";

type InsightsLoadingProps = {
  label?: string;
};

export function InsightsLoading({ label = "Fetching insights from GitHub…" }: InsightsLoadingProps) {
  return (
    <Fade in timeout={300}>
      <Box sx={{ p: 2 }}>
        <Stack spacing={2}>
          <Typography variant="body2" color="text.secondary">
            {label}
          </Typography>
          <LinearProgress />
          <Box sx={{ display: "flex", gap: 2, flexWrap: "wrap" }}>
            <Skeleton variant="circular" width={180} height={180} animation="wave" />
            <Box sx={{ flex: 1, minWidth: 200 }}>
              <Skeleton variant="rounded" height={180} animation="wave" />
            </Box>
          </Box>
          <Skeleton variant="rounded" height={120} animation="wave" />
        </Stack>
      </Box>
    </Fade>
  );
}
