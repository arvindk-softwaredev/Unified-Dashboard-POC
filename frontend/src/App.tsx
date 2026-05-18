import "@fontsource/roboto/300.css";
import "@fontsource/roboto/400.css";
import "@fontsource/roboto/500.css";
import "@fontsource/roboto/700.css";
import { Dashboard } from "./dashboard/Dashboard";
import { AiModeProvider } from "./theme/AiModeContext";
import { ColorModeProvider } from "./theme/ColorModeContext";
import "./index.css";

export function App() {
  return (
    <ColorModeProvider>
      <AiModeProvider>
        <Dashboard />
      </AiModeProvider>
    </ColorModeProvider>
  );
}

export default App;
