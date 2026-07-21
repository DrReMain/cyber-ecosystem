import { useRouter } from "@tanstack/react-router";

export function ErrorPage({
  error,
  showDetails,
}: Readonly<{ error: Error; showDetails?: boolean }>) {
  const router = useRouter();
  return (
    <div>
      <h1>500</h1>
      <p>Something went wrong</p>
      <button onClick={() => router.navigate({ to: "/" })} type="button">
        Back Home
      </button>
      <button onClick={() => router.invalidate()} type="button">
        Retry
      </button>
      {showDetails ? <pre>{String(error)}</pre> : null}
    </div>
  );
}
