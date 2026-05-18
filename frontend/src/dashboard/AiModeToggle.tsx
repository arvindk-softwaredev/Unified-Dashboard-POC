import AutoAwesomeIcon from "@mui/icons-material/AutoAwesome";
import ToggleButton from "@mui/material/ToggleButton";
import Tooltip from "@mui/material/Tooltip";
import { useAiMode } from "../theme/AiModeContext";

export function AiModeToggle() {
  const { aiMode, toggleAiMode } = useAiMode();

  return (
    <Tooltip title={aiMode ? "Turn off AI Mode" : "Turn on AI Mode — animated gradient sweep"}>
      <ToggleButton
        value="ai"
        selected={aiMode}
        onClick={toggleAiMode}
        size="small"
        sx={{
          ml: 1,
          textTransform: "none",
          gap: 0.75,
          "&.Mui-selected": {
            background: theme =>
              `linear-gradient(90deg, ${theme.palette.primary.main}, #9c27b0)`,
            color: "#fff",
            "&:hover": {
              background: theme =>
                `linear-gradient(90deg, ${theme.palette.primary.dark}, #7b1fa2)`,
            },
          },
        }}
      >
        <AutoAwesomeIcon sx={{ fontSize: 18 }} />
        AI Mode
      </ToggleButton>
    </Tooltip>
  );
}
