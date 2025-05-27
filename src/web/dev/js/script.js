document.addEventListener('DOMContentLoaded', function () {
  const conn = initWebSocket();
  const msg = document.getElementById("msg");
  const log = document.getElementById("log");
  const form = document.getElementById("form");
  const connectionStatus = document.getElementById("connection-status");

  function initWebSocket() {
    if (!window["WebSocket"]) {
      showSystemMessage("Your browser does not support WebSockets.");
      return null;
    }

    const conn = new WebSocket("ws://" + document.location.host + "/ws");

    conn.onopen = function () {
      connectionStatus.textContent = "Connected";
      connectionStatus.className = "connection-status connected";
      showSystemMessage("Connection established");
    };

    conn.onclose = function () {
      connectionStatus.textContent = "Disconnected";
      connectionStatus.className = "connection-status disconnected";
      showSystemMessage("Connection closed");
    };

    conn.onmessage = function (evt) {
      const messages = evt.data.split('\n');
      for (let i = 0; i < messages.length; i++) {
        addMessage(messages[i], "user-message");
      }
    };

    return conn;
  }

  function showSystemMessage(text) {
    addMessage(text, "system-message");
  }

  function addMessage(text, className) {
    const item = document.createElement("div");
    item.className = `message ${className}`;

    if (className.includes("user-message")) {
      const parts = text.split(/:(.*)/s); // Splits at the first colon
      const senderName = parts[0];
      const actualMessage = parts[1] || ""; // Handle if message is empty after colon

      item.innerHTML = ''; // Clear any existing text content

      const senderSpan = document.createElement('span');
      senderSpan.className = 'message-sender';
      senderSpan.textContent = senderName + ': '; // Add colon and space for display

      const messageSpan = document.createElement('span');
      messageSpan.className = 'message-text';
      messageSpan.textContent = actualMessage;

      item.appendChild(senderSpan);
      item.appendChild(messageSpan);
    } else {
      item.textContent = text;
    }

    appendToLog(item);
  }

  function appendToLog(item) {
    const doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
    log.appendChild(item);
    if (doScroll) {
      log.scrollTop = log.scrollHeight - log.clientHeight;
    }
  }

  // Form submission handler
  form.onsubmit = function () {
    if (!conn) {
      return false;
    }
    if (!msg.value) {
      return false;
    }

    conn.send(msg.value);
    msg.value = "";
    return false;
  };

  // Focus input field automatically
  msg.focus();
});