export interface Pushable<T> extends AsyncIterable<T> {
  push(item: T): void;
  end(): void;
}

export function createPushable<T>(): Pushable<T> {
  const queue: T[] = [];
  let notify: (() => void) | null = null;
  let ended = false;

  const wake = () => {
    const n = notify;
    notify = null;
    n?.();
  };

  return {
    push(item) {
      if (ended) return;
      queue.push(item);
      wake();
    },
    end() {
      ended = true;
      wake();
    },
    [Symbol.asyncIterator]() {
      return {
        next: (): Promise<IteratorResult<T>> =>
          new Promise((resolve) => {
            const attempt = () => {
              if (queue.length > 0) {
                resolve({ value: queue.shift() as T, done: false });
              } else if (ended) {
                resolve({ value: undefined, done: true });
              } else {
                notify = attempt;
              }
            };
            attempt();
          }),
      };
    },
  };
}
