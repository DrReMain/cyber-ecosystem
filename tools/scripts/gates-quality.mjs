// gates:quality — repo lint, then affected check+typecheck against the
// lane-aware base (see affected-base.mjs for the routing law).
import { nx } from './lib.mjs';
import { affectedBase } from './affected-base.mjs';

nx(['run', 'tools:go:lint']);
nx(['affected', '-t', 'check', 'typecheck', '--base', affectedBase(), '--parallel', '3']);
