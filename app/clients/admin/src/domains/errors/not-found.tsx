import { useRouter } from "@tanstack/react-router";

export function NotFound() {
  const router = useRouter();
  return (
    <div>
      <h1>404</h1>
      <button onClick={() => router.navigate({ to: "/" })} type="button">
        Back Home
      </button>
    </div>
  );
}
