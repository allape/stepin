
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function stringify(e?: Error | any): string {
  return e?.message || `${e}`;
}
