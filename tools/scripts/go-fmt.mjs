// tools:go:fmt — gofmt check without rewriting: unformatted files fail the
// gate naming themselves; the fix is a separate deliberate action.
import { spawnSync } from 'node:child_process';

const r = spawnSync('gofmt', ['-l', 'app', 'shared-go'], { encoding: 'utf8' });
if (r.error) {
  console.error(`gofmt not available: ${r.error.message}`);
  process.exit(1);
}
if (r.stdout.trim()) {
  console.error(r.stdout.trim());
  console.error('gofmt needed (run: gofmt -w the files above)');
  process.exit(1);
}
