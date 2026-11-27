import { useCallback, useRef } from "react";

// The parent owns one dirty flag. A cancelled navigation must not clear it.
export function usePublishedDraftGuard() {
  const dirty = useRef(false);
  const onDirtyChange = useCallback((value: boolean) => { dirty.current = value; }, []);
  const allowDiscard = useCallback(() => {
    if (dirty.current && !window.confirm("当前商品草稿有未保存修改，确定放弃修改并离开？")) return false;
    dirty.current = false;
    return true;
  }, []);
  const hasUnsaved = useCallback(() => dirty.current, []);
  return { onDirtyChange, allowDiscard, hasUnsaved };
}
