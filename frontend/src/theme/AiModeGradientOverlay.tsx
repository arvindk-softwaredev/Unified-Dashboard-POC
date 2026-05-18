import Box from "@mui/material/Box";
import { useEffect, useRef, useState } from "react";

type Phase = "hidden" | "sweep-in" | "stable" | "sweep-out";

type AiModeGradientOverlayProps = {
  enabled: boolean;
};

export function AiModeGradientOverlay({ enabled }: AiModeGradientOverlayProps) {
  const [phase, setPhase] = useState<Phase>(() => (enabled ? "stable" : "hidden"));
  const wasEnabled = useRef(enabled);

  useEffect(() => {
    if (enabled && !wasEnabled.current) {
      setPhase("sweep-in");
    } else if (!enabled && wasEnabled.current) {
      setPhase("sweep-out");
    }
    wasEnabled.current = enabled;
  }, [enabled]);

  const handleBeamEnd = () => {
    if (phase === "sweep-in") setPhase("stable");
    if (phase === "sweep-out") setPhase("hidden");
  };

  if (phase === "hidden") return null;

  const showBeam = phase === "sweep-in" || phase === "sweep-out";
  const showBackground = phase === "stable" || phase === "sweep-out";
  const backgroundStable = phase === "stable";
  const backgroundSweepOut = phase === "sweep-out";

  return (
    <Box
      className="ai-mode-overlay"
      aria-hidden
      sx={{
        position: "fixed",
        inset: 0,
        pointerEvents: "none",
        zIndex: 1300,
        overflow: "hidden",
      }}
    >
      {showBackground && (
        <Box
          className={[
            "ai-mode-background",
            backgroundStable && "ai-mode-background--visible",
            backgroundSweepOut && "ai-mode-background--sweep-out",
          ]
            .filter(Boolean)
            .join(" ")}
        />
      )}
      {showBeam && (
        <Box
          className={[
            "ai-mode-gradient-beam",
            phase === "sweep-out" && "ai-mode-gradient-beam--reverse",
          ]
            .filter(Boolean)
            .join(" ")}
          onAnimationEnd={handleBeamEnd}
        />
      )}
    </Box>
  );
}
