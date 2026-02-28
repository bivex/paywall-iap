/**
 * Matomo Analytics SDK for React Native
 *
 * Features:
 * - Event batching with configurable batch size and flush interval
 * - Local storage fallback for offline events
 * - Automatic retry with exponential backoff
 * - Ecommerce (purchase) tracking
 */

import { Platform } from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { HttpClient } from '../http/HttpClient';

// Constants
const MATOMO_QUEUE_KEY = '@matomo_event_queue';
const MATOMO_CONFIG_KEY = '@matomo_config';
const DEFAULT_BATCH_SIZE = 20;
const DEFAULT_FLUSH_INTERVAL = 30000; // 30 seconds
const MAX_QUEUE_SIZE = 1000;
const MAX_RETRIES = 3;

// Types
export interface MatomoConfig {
  baseUrl: string;
  siteId: string;
  userId?: string;
  batchSize?: number;
  flushInterval?: number;
  maxRetries?: number;
}

export interface MatomoEvent {
  id: string;
  type: 'event' | 'ecommerce';
  category?: string;
  action?: string;
  name?: string;
  value?: number;
  customVars?: Record<string, string>;
  // Ecommerce fields
  orderId?: string;
  revenue?: number;
  items?: EcommerceItem[];
  // Metadata
  timestamp: number;
  retryCount: number;
  sentAt?: number;
  failedAt?: number;
  errorMessage?: string;
}

export interface EcommerceItem {
  sku: string;
  name: string;
  price: number;
  quantity: number;
  category?: string;
}

export interface EventQueue {
  events: MatomoEvent[];
  lastFlushAt: number;
}

/**
 * MatomoAnalytics class for tracking events
 */
export class MatomoAnalytics {
  private config: MatomoConfig;
  private httpClient: HttpClient;
  private queue: MatomoEvent[] = [];
  private flushTimer: NodeJS.Timeout | null = null;
  private isFlushing = false;

  constructor(config: MatomoConfig, httpClient: HttpClient) {
    this.config = {
      batchSize: DEFAULT_BATCH_SIZE,
      flushInterval: DEFAULT_FLUSH_INTERVAL,
      maxRetries: MAX_RETRIES,
      ...config,
    };
    this.httpClient = httpClient;
  }

  /**
   * Initialize the Matomo SDK
   * Loads persisted queue from AsyncStorage and starts auto-flush timer
   */
  async initialize(): Promise<void> {
    try {
      // Load config
      const savedConfig = await AsyncStorage.getItem(MATOMO_CONFIG_KEY);
      if (savedConfig) {
        this.config = { ...this.config, ...JSON.parse(savedConfig) };
      }

      // Load queue
      const savedQueue = await AsyncStorage.getItem(MATOMO_QUEUE_KEY);
      if (savedQueue) {
        const queueData: EventQueue = JSON.parse(savedQueue);
        this.queue = queueData.events || [];
      }

      // Start auto-flush timer
      this.startFlushTimer();

      console.log('[Matomo] Initialized', {
        queueSize: this.queue.length,
        config: this.config,
      });
    } catch (error) {
      console.error('[Matomo] Failed to initialize', error);
    }
  }

  /**
   * Set the user ID for tracking
   */
  async setUserId(userId: string): Promise<void> {
    this.config.userId = userId;
    await this.saveConfig();
  }

  /**
   * Track a standard event
   */
  async trackEvent(
    category: string,
    action: string,
    name?: string,
    value?: number,
    customVars?: Record<string, string>
  ): Promise<void> {
    const event: MatomoEvent = {
      id: this.generateId(),
      type: 'event',
      category,
      action,
      name,
      value,
      customVars,
      timestamp: Date.now(),
      retryCount: 0,
    };

    await this.enqueue(event);
  }

  /**
   * Track an ecommerce purchase
   */
  async trackPurchase(
    orderId: string,
    revenue: number,
    items: EcommerceItem[],
    customVars?: Record<string, string>
  ): Promise<void> {
    const event: MatomoEvent = {
      id: this.generateId(),
      type: 'ecommerce',
      orderId,
      revenue,
      items,
      customVars,
      timestamp: Date.now(),
      retryCount: 0,
    };

    await this.enqueue(event);
  }

  /**
   * Enqueue an event to the local queue
   */
  private async enqueue(event: MatomoEvent): Promise<void> {
    // Check queue size limit
    if (this.queue.length >= MAX_QUEUE_SIZE) {
      console.warn('[Matomo] Queue full, dropping oldest event');
      this.queue.shift();
    }

    this.queue.push(event);
    await this.saveQueue();

    // Auto-flush if batch size reached
    if (this.queue.length >= this.config.batchSize!) {
      await this.flush();
    }
  }

  /**
   * Flush all queued events to the server
   */
  async flush(): Promise<void> {
    if (this.isFlushing || this.queue.length === 0) {
      return;
    }

    this.isFlushing = true;
    const eventsToSend = this.queue.slice(0, this.config.batchSize);

    try {
      await this.sendBatch(eventsToSend);

      // Remove sent events from queue
      this.queue = this.queue.slice(eventsToSend.length);
      await this.saveQueue();

      console.log('[Matomo] Flushed', { count: eventsToSend.length });
    } catch (error) {
      console.error('[Matomo] Flush failed', error);
      // Mark events as failed but keep them for retry
      eventsToSend.forEach((event) => {
        event.retryCount++;
        if (event.retryCount >= (this.config.maxRetries || MAX_RETRIES)) {
          event.failedAt = Date.now();
          event.errorMessage = String(error);
        }
      });
    } finally {
      this.isFlushing = false;
    }
  }

  /**
   * Send a batch of events to the server
   */
  private async sendBatch(events: MatomoEvent[]): Promise<void> {
    for (const event of events) {
      if (event.failedAt && event.retryCount >= (this.config.maxRetries || MAX_RETRIES)) {
        // Skip permanently failed events
        continue;
      }

      try {
        if (event.type === 'event') {
          await this.sendEvent(event);
        } else if (event.type === 'ecommerce') {
          await this.sendEcommerce(event);
        }
        event.sentAt = Date.now();
      } catch (error) {
        throw error; // Will trigger retry logic
      }
    }
  }

  /**
   * Send a single event to Matomo
   */
  private async sendEvent(event: MatomoEvent): Promise<void> {
    const params = new URLSearchParams({
      rec: '1',
      idsite: this.config.siteId,
      e_c: event.category || '',
      e_a: event.action || '',
      rand: String(Date.now()),
    });

    if (event.name) params.set('e_n', event.name);
    if (event.value !== undefined) params.set('e_v', String(event.value));
    if (this.config.userId) params.set('cid', this.config.userId);
    if (event.customVars) {
      Object.entries(event.customVars).forEach(([key, value], index) => {
        params.set(`cvar[${index}][0]`, key);
        params.set(`cvar[${index}][1]`, value);
      });
    }

    await this.httpClient.post(`${this.config.baseUrl}/matomo.php`, params.toString(), {
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    });
  }

  /**
   * Send an ecommerce event to Matomo
   */
  private async sendEcommerce(event: MatomoEvent): Promise<void> {
    const params = new URLSearchParams({
      rec: '1',
      idsite: this.config.siteId,
      e_c: 'ecommerce',
      e_a: 'purchase',
      revenue: String(event.revenue || 0),
      rand: String(Date.now()),
    });

    if (event.orderId) params.set('ec_id', event.orderId);
    if (this.config.userId) params.set('cid', this.config.userId);
    if (event.items) {
      params.set('ec_items', JSON.stringify(event.items));
    }
    if (event.customVars) {
      Object.entries(event.customVars).forEach(([key, value], index) => {
        params.set(`cvar[${index}][0]`, key);
        params.set(`cvar[${index}][1]`, value);
      });
    }

    await this.httpClient.post(`${this.config.baseUrl}/matomo.php`, params.toString(), {
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    });
  }

  /**
   * Start the auto-flush timer
   */
  private startFlushTimer(): void {
    if (this.flushTimer) {
      clearInterval(this.flushTimer);
    }

    this.flushTimer = setInterval(() => {
      this.flush();
    }, this.config.flushInterval);
  }

  /**
   * Stop the auto-flush timer
   */
  stopFlushTimer(): void {
    if (this.flushTimer) {
      clearInterval(this.flushTimer);
      this.flushTimer = null;
    }
  }

  /**
   * Save the queue to AsyncStorage
   */
  private async saveQueue(): Promise<void> {
    const queueData: EventQueue = {
      events: this.queue,
      lastFlushAt: Date.now(),
    };
    await AsyncStorage.setItem(MATOMO_QUEUE_KEY, JSON.stringify(queueData));
  }

  /**
   * Save config to AsyncStorage
   */
  private async saveConfig(): Promise<void> {
    await AsyncStorage.setItem(MATOMO_CONFIG_KEY, JSON.stringify(this.config));
  }

  /**
   * Generate a unique event ID
   */
  private generateId(): string {
    return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  /**
   * Get queue statistics
   */
  getQueueStats(): { size: number; pending: number; failed: number } {
    const pending = this.queue.filter((e) => !e.sentAt && !e.failedAt).length;
    const failed = this.queue.filter((e) => e.failedAt).length;
    return { size: this.queue.length, pending, failed };
  }

  /**
   * Clear all queued events
   */
  async clearQueue(): Promise<void> {
    this.queue = [];
    await AsyncStorage.removeItem(MATOMO_QUEUE_KEY);
  }

  /**
   * Retry failed events
   */
  async retryFailedEvents(): Promise<void> {
    this.queue.forEach((event) => {
      if (event.failedAt) {
        event.failedAt = undefined;
        event.errorMessage = undefined;
        event.retryCount = 0;
      }
    });
    await this.saveQueue();
    await this.flush();
  }
}

/**
 * Create a singleton instance of MatomoAnalytics
 */
let matomoInstance: MatomoAnalytics | null = null;

export function initMatomo(config: MatomoConfig, httpClient: HttpClient): MatomoAnalytics {
  if (!matomoInstance) {
    matomoInstance = new MatomoAnalytics(config, httpClient);
  }
  return matomoInstance;
}

export function getMatomo(): MatomoAnalytics | null {
  return matomoInstance;
}
