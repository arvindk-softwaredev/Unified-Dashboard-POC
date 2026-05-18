import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { AiModeGradientOverlay } from "./AiModeGradientOverlay";

type AiModeContextValue = {
  aiMode: boolean;
  toggleAiMode: () => void;
};

const AiModeContext = createContext<AiModeContextValue | undefined>(undefined);

export function useAiMode() {
  const ctx = useContext(AiModeContext);
  if (!ctx) throw new Error("useAiMode must be used within AiModeProvider");
  return ctx;
}

export function AiModeProvider({ children }: { children: ReactNode }) {
  const [aiMode, setAiMode] = useState(() => localStorage.getItem("aiMode") === "true");

  const toggleAiMode = useCallback(() => {
    setAiMode(prev => {
      const next = !prev;
      localStorage.setItem("aiMode", String(next));
      return next;
    });
  }, []);

  useEffect(() => {
    document.documentElement.classList.toggle("ai-mode-active", aiMode);
    return () => document.documentElement.classList.remove("ai-mode-active");
  }, [aiMode]);

  const value = useMemo(() => ({ aiMode, toggleAiMode }), [aiMode, toggleAiMode]);

  return (
    <AiModeContext.Provider value={value}>
      {children}
      <AiModeGradientOverlay enabled={aiMode} />
    </AiModeContext.Provider>
  );
}
