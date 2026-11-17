// gates:specs — affected tests against the lane-aware base (see
// affected-base.mjs for the routing law).
import { nx } from './lib.mjs';
import { affectedBase } from './affected-base.mjs';

nx(['affected', '-t', 'test', '--base', affectedBase(), '--parallel', '3']);
