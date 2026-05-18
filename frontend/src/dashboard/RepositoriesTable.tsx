import KeyboardArrowDownIcon from "@mui/icons-material/KeyboardArrowDown";
import KeyboardArrowUpIcon from "@mui/icons-material/KeyboardArrowUp";
import RefreshIcon from "@mui/icons-material/Refresh";
import Chip from "@mui/material/Chip";
import Collapse from "@mui/material/Collapse";
import IconButton from "@mui/material/IconButton";
import Paper from "@mui/material/Paper";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Tooltip from "@mui/material/Tooltip";
import Typography from "@mui/material/Typography";
import Box from "@mui/material/Box";
import { Fragment, useEffect, useRef, useState } from "react";
import type { Repository } from "../api/client";
import { RepositoryInsightPanel } from "./RepositoryInsightPanel";

type RepositoriesTableProps = {
  repositories: Repository[];
};

export function RepositoriesTable({ repositories }: RepositoriesTableProps) {
  const [expandedId, setExpandedId] = useState<number | null>(null);
  const [refreshByRepo, setRefreshByRepo] = useState<Record<number, number>>({});
  const rowRefs = useRef<Map<number, HTMLTableRowElement>>(new Map());

  const setRowRef = (id: number) => (el: HTMLTableRowElement | null) => {
    if (el) rowRefs.current.set(id, el);
    else rowRefs.current.delete(id);
  };

  useEffect(() => {
    if (expandedId == null) return;
    const frame = requestAnimationFrame(() => {
      rowRefs.current.get(expandedId)?.scrollIntoView({
        behavior: "smooth",
        block: "start",
      });
    });
    return () => cancelAnimationFrame(frame);
  }, [expandedId]);

  const toggle = (id: number) => {
    setExpandedId(prev => (prev === id ? null : id));
  };

  const handleRefresh = (repo: Repository, e: React.MouseEvent) => {
    e.stopPropagation();
    setExpandedId(repo.id);
    setRefreshByRepo(prev => ({
      ...prev,
      [repo.id]: (prev[repo.id] ?? 0) + 1,
    }));
  };

  return (
    <Paper variant="outlined" sx={{ display: "flex", flexDirection: "column", minHeight: 0, flex: 1 }}>
      <TableContainer sx={{ flex: 1, overflow: "auto" }}>
        <Table size="small" stickyHeader>
          <TableHead>
            <TableRow>
              <TableCell width={48} />
              <TableCell>Repository</TableCell>
              <TableCell>Description</TableCell>
              <TableCell>Language</TableCell>
              <TableCell align="right">Stars</TableCell>
              <TableCell align="right">Open issues</TableCell>
              <TableCell width={52} align="center">
                <Tooltip title="Refresh fetches latest GitHub data and bypasses the 15-minute cache">
                  <span>↻</span>
                </Tooltip>
              </TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {repositories.map(repo => {
              const open = expandedId === repo.id;
              const refreshTrigger = refreshByRepo[repo.id] ?? 0;
              return (
                <Fragment key={repo.id}>
                  <TableRow
                    ref={setRowRef(repo.id)}
                    hover
                    sx={{ cursor: "pointer", scrollMarginTop: 8 }}
                    onClick={() => toggle(repo.id)}
                  >
                    <TableCell>
                      <IconButton size="small" aria-label={open ? "collapse" : "expand"}>
                        {open ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
                      </IconButton>
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2" fontWeight={600}>
                        {repo.name}
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        {repo.full_name}
                      </Typography>
                    </TableCell>
                    <TableCell sx={{ maxWidth: 360 }}>
                      <Typography variant="body2" noWrap title={repo.description ?? ""}>
                        {repo.description || "—"}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      {repo.language ? <Chip label={repo.language} size="small" /> : "—"}
                    </TableCell>
                    <TableCell align="right">{repo.stargazers_count.toLocaleString()}</TableCell>
                    <TableCell align="right">{repo.open_issues_count.toLocaleString()}</TableCell>
                    <TableCell align="center" onClick={e => e.stopPropagation()}>
                      <Tooltip title="Refresh insights from GitHub (bypasses cache, shows loading animation)">
                        <IconButton
                          size="small"
                          aria-label="Refresh repository insights"
                          onClick={e => handleRefresh(repo, e)}
                        >
                          <RefreshIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    </TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell colSpan={7} sx={{ p: 0, borderBottom: open ? undefined : 0 }}>
                      <Collapse in={open} timeout="auto" unmountOnExit>
                        <Box sx={{ maxHeight: "min(70vh, 560px)", overflow: "auto" }}>
                          {open && (
                            <RepositoryInsightPanel repo={repo} refreshTrigger={refreshTrigger} />
                          )}
                        </Box>
                      </Collapse>
                    </TableCell>
                  </TableRow>
                </Fragment>
              );
            })}
          </TableBody>
        </Table>
      </TableContainer>
    </Paper>
  );
}
