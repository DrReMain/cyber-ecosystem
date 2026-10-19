export interface Channel<T> {
  publish(message: T): void;
  subscribe(handler: (message: T) => void): () => void;
  close(): void;
}

export function createChannel<T>(name: string): Channel<T> {
  if (typeof BroadcastChannel !== "undefined") {
    const channel = new BroadcastChannel(name);
    return {
      publish: (message) => channel.postMessage(message),
      subscribe: (handler) => {
        channel.onmessage = (event) => handler(event.data as T);
        return () => {
          channel.onmessage = null;
        };
      },
      close: () => channel.close(),
    };
  }
  let seq = 0;
  return {
    publish: (message) => localStorage.setItem(name, JSON.stringify({ n: ++seq, message })),
    subscribe: (handler) => {
      const onStorage = (event: StorageEvent) => {
        if (event.key !== name || event.newValue === null) {
          return;
        }
        try {
          handler((JSON.parse(event.newValue) as { message: T }).message);
        } catch {
          // Foreign payloads under the same key are ignored.
        }
      };
      window.addEventListener("storage", onStorage);
      return () => window.removeEventListener("storage", onStorage);
    },
    close: () => {},
  };
}
