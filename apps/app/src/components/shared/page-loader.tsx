import { Spinner } from "./spinner";

/** Centered loading spinner for page-level content areas. */
export function PageLoader() {
  return (
    <div className="flex flex-1 items-center justify-center py-20">
      <Spinner />
    </div>
  );
}
