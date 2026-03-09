export function nFormatter(num: number, digits: number = 1): string {
  const formatter = new Intl.NumberFormat("en-US", {
    notation: "compact",
    maximumFractionDigits: digits,
  });
  return formatter.format(num);
}
