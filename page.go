package ssehub

import (
	"fmt"
	"net/http"
)

// TODO 19.05.2025 autoscroll

func (hub *Hub) pageHandler(w http.ResponseWriter, r *http.Request) {
	html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<body style="background-color: #333; color: #eee">
			<pre id="logs" style="position: fixed; top: 0; bottom: 0; left: 0; right: 0; padding: 0 8px; overflow: auto"></pre>
			<script>
				const logBox = document.getElementById('logs');
				const es = new EventSource('%s');

				es.onmessage = function(event) {
					const msg = event.data.trim();

					if (!msg.length) return;

					logBox.textContent += msg + "\n";
				};
			</script>
		</body>
		</html>
	`, r.RequestURI)

	w.Write([]byte(html))
}
