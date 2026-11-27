import { createContext, useContext } from "react";
import type { createFeedback } from "./operationFeedback";

export const OperationFeedbackContext = createContext<ReturnType<typeof createFeedback> | null>(null);

export function useOperationFeedback() {
  const feedback = useContext(OperationFeedbackContext);
  if (!feedback) throw new Error("Operation feedback requires an authenticated workspace");
  return feedback;
}
