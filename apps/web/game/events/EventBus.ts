type EventCallback = (...args: unknown[]) => void;

class EventBus {
  private listeners: Map<string, EventCallback[]> = new Map();

  on(event: string, callback: EventCallback): void {
    const existing = this.listeners.get(event) || [];
    existing.push(callback);
    this.listeners.set(event, existing);
  }

  off(event: string, callback: EventCallback): void {
    const existing = this.listeners.get(event);
    if (!existing) return;
    this.listeners.set(
      event,
      existing.filter((cb) => cb !== callback)
    );
  }

  emit(event: string, ...args: unknown[]): void {
    const existing = this.listeners.get(event);
    if (!existing) return;
    existing.forEach((cb) => cb(...args));
  }

  removeAllListeners(event?: string): void {
    if (event) {
      this.listeners.delete(event);
    } else {
      this.listeners.clear();
    }
  }
}

const eventBus = new EventBus();
export default eventBus;
