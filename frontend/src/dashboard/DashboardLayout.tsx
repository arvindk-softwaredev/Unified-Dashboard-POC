import AppBar from "@mui/material/AppBar";
import Box from "@mui/material/Box";
import Container from "@mui/material/Container";
import Toolbar from "@mui/material/Toolbar";
import Typography from "@mui/material/Typography";
import { alpha } from "@mui/material/styles";
import type { ReactNode } from "react";
import { AiModeToggle } from "./AiModeToggle";
import { ThemeToggle } from "./ThemeToggle";

type DashboardLayoutProps = {
  children: ReactNode;
};

export function DashboardLayout({ children }: DashboardLayoutProps) {
  return (
    <Box sx={{ display: "flex", flexDirection: "column", minHeight: "100vh" }}>
      <AppBar position="sticky" elevation={0} color="default" sx={{ borderBottom: 1, borderColor: "divider" }}>
        <Toolbar>
          <Typography variant="h6" sx={{ flexGrow: 1, fontWeight: 700 }}>
            Tekton Unified Dashboard
          </Typography>
          <ThemeToggle />
          <AiModeToggle />
        </Toolbar>
      </AppBar>
      <Box
        component="main"
        sx={theme => ({
          flexGrow: 1,
          backgroundColor: theme.vars
            ? `rgba(${theme.vars.palette.background.defaultChannel} / 1)`
            : alpha(theme.palette.background.default, 1),
        })}
      >
        <Container maxWidth="xl" sx={{ py: { xs: 2, md: 3 }, height: "100%", display: "flex", flexDirection: "column" }}>
          {children}
        </Container>
      </Box>
    </Box>
  );
}
