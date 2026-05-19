import Box from "@mui/material/Box";
import Paper from "@mui/material/Paper";
import Skeleton from "@mui/material/Skeleton";
import Stack from "@mui/material/Stack";

export function RepositoryCardSkeleton() {
  return (
    <Paper variant="outlined" sx={{ borderRadius: 2, overflow: "hidden" }}>
      <Box sx={{ p: 2 }}>
        <Stack direction="row" spacing={1.5} sx={{ alignItems: "flex-start" }}>
          <Skeleton variant="circular" width={32} height={32} animation="wave" sx={{ mt: 0.25 }} />
          <Skeleton variant="circular" width={40} height={40} animation="wave" />
          <Box sx={{ flex: 1, minWidth: 0 }}>
            <Stack
              direction={{ xs: "column", sm: "row" }}
              spacing={{ xs: 1, sm: 2 }}
              sx={{ alignItems: { xs: "flex-start", sm: "center" }, justifyContent: "space-between" }}
            >
              <Box sx={{ width: "100%", maxWidth: 280 }}>
                <Skeleton variant="text" width="70%" height={26} animation="wave" />
                <Skeleton variant="text" width="50%" height={16} animation="wave" sx={{ mt: 0.5 }} />
              </Box>
              <Stack direction="row" spacing={0.75} useFlexGap sx={{ flexWrap: "wrap" }}>
                <Skeleton variant="rounded" width={56} height={28} animation="wave" />
                <Skeleton variant="rounded" width={48} height={28} animation="wave" />
                <Skeleton variant="rounded" width={52} height={28} animation="wave" />
                <Skeleton variant="rounded" width={72} height={28} animation="wave" />
                <Skeleton variant="circular" width={32} height={32} animation="wave" />
              </Stack>
            </Stack>
            <Skeleton variant="text" width="95%" height={18} animation="wave" sx={{ mt: 1.25 }} />
            <Skeleton variant="text" width="80%" height={18} animation="wave" sx={{ mt: 0.5 }} />
          </Box>
        </Stack>
      </Box>
    </Paper>
  );
}
