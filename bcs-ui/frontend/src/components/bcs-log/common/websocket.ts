export default class BCSWebSocket {
  timeout = 6000;
  timeoutObj: any = null;
  lockReconnect = false;
  lockTimeoutObj: any = null;
  ws!: WebSocket;

  constructor(url: string) {
    this.createWebSocket(url);
  }

  createWebSocket(url: string) {
    this.ws = new WebSocket(url);
    this.ws.addEventListener('open', () => {
      console.log('WebSocket Open');
    });
    this.ws.addEventListener('close', () => {
      console.log('WebSocket close');
    });
    this.ws.addEventListener('message', () => {
      console.log('WebSocket message');
    });
    this.ws.addEventListener('error', () => {
      console.log('WebSocket error');
    });
  }

  reconnect(url: string) {
    if (this.lockReconnect) return;

    this.lockReconnect = true;

    this.lockTimeoutObj && clearTimeout(this.lockTimeoutObj);

    this.lockTimeoutObj = setTimeout(() => {
      this.createWebSocket(url);
      this.lockReconnect = false;
    }, 2000);
  }

  reset() {
    clearTimeout(this.timeoutObj);
    this.start();
  }

  start() {
    this.timeoutObj = setTimeout(() => {
      this.ws.send('heart beat');
    }, this.timeout);
  }
}
