#include <arpa/inet.h>
#include <ctype.h>
#include <errno.h>
#include <netinet/in.h>
#include <signal.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <time.h>
#include <unistd.h>

#define HEADER_LIMIT 16384
#define CHUNK_SIZE 4096

struct request {
  char method[16];
  char path[256];
  char content_type[128];
  char user_agent[256];
  char *body;
  size_t body_len;
  int error_status;
};

static const char *getenv_or(const char *key, const char *fallback) {
  const char *value = getenv(key);
  return (value && *value) ? value : fallback;
}

static size_t getenv_size(const char *key, size_t fallback) {
  const char *value = getenv(key);
  if (!value || !*value) return fallback;
  char *end = NULL;
  unsigned long long parsed = strtoull(value, &end, 10);
  return (end && *end == '\0') ? (size_t)parsed : fallback;
}

static void trim(char *s) {
  size_t len = strlen(s);
  while (len > 0 && (s[len - 1] == ' ' || s[len - 1] == '\t' || s[len - 1] == '\r' || s[len - 1] == '\n')) {
    s[--len] = '\0';
  }
  size_t start = 0;
  while (s[start] == ' ' || s[start] == '\t' || s[start] == '\r' || s[start] == '\n') start++;
  if (start > 0) memmove(s, s + start, len - start + 1);
}

static void lower_ascii(char *s) {
  for (; *s; ++s) *s = (char)tolower((unsigned char)*s);
}

static void json_escape_file(FILE *out, const char *data, size_t len) {
  for (size_t i = 0; i < len; ++i) {
    unsigned char c = (unsigned char)data[i];
    switch (c) {
      case '"': fputs("\\\"", out); break;
      case '\\': fputs("\\\\", out); break;
      case '\b': fputs("\\b", out); break;
      case '\f': fputs("\\f", out); break;
      case '\n': fputs("\\n", out); break;
      case '\r': fputs("\\r", out); break;
      case '\t': fputs("\\t", out); break;
      default:
        if (c < 0x20) fprintf(out, "\\u%04x", c);
        else fputc((int)c, out);
    }
  }
}

static void now_iso8601(char *buf, size_t size) {
  time_t now = time(NULL);
  struct tm tm_utc;
  gmtime_r(&now, &tm_utc);
  strftime(buf, size, "%Y-%m-%dT%H:%M:%SZ", &tm_utc);
}

static int ensure_parent_dirs(const char *path) {
  char tmp[512];
  size_t len = strlen(path);
  if (len >= sizeof(tmp)) return 0;
  memcpy(tmp, path, len + 1);
  for (size_t i = 1; i < len; ++i) {
    if (tmp[i] == '/') {
      tmp[i] = '\0';
      if (mkdir(tmp, 0755) != 0 && errno != EEXIST) return 0;
      tmp[i] = '/';
    }
  }
  return 1;
}

static void send_response(int fd, int status, const char *reason, const char *body, const char *content_type) {
  char head[512];
  int head_len = snprintf(head, sizeof(head),
    "HTTP/1.1 %d %s\r\nContent-Type: %s\r\nContent-Length: %zu\r\nConnection: close\r\n\r\n",
    status, reason, content_type ? content_type : "application/json", strlen(body));
  if (head_len > 0) send(fd, head, (size_t)head_len, 0);
  send(fd, body, strlen(body), 0);
}

static void remote_addr_str(const struct sockaddr_storage *addr, char *buf, size_t size) {
  buf[0] = '\0';
  if (addr->ss_family == AF_INET) {
    const struct sockaddr_in *a = (const struct sockaddr_in *)addr;
    inet_ntop(AF_INET, &a->sin_addr, buf, (socklen_t)size);
  } else if (addr->ss_family == AF_INET6) {
    const struct sockaddr_in6 *a = (const struct sockaddr_in6 *)addr;
    inet_ntop(AF_INET6, &a->sin6_addr, buf, (socklen_t)size);
  }
  if (!buf[0]) snprintf(buf, size, "unknown");
}

static void free_request(struct request *req) {
  free(req->body);
  req->body = NULL;
}

static void set_header_value(char *dst, size_t dst_size, const char *value) {
  snprintf(dst, dst_size, "%s", value ? value : "");
  trim(dst);
}

static struct request read_request(int fd, size_t max_body_bytes) {
  struct request req;
  memset(&req, 0, sizeof(req));
  char header_buf[HEADER_LIMIT + 1];
  size_t total = 0;
  ssize_t n;
  size_t header_end = 0;
  int found = 0;

  while (!found) {
    if (total == HEADER_LIMIT) { req.error_status = 431; return req; }
    n = recv(fd, header_buf + total, HEADER_LIMIT - total, 0);
    if (n <= 0) { req.error_status = 400; return req; }
    total += (size_t)n;
    header_buf[total] = '\0';
    char *p = strstr(header_buf, "\r\n\r\n");
    if (p) { header_end = (size_t)(p - header_buf); found = 1; }
  }

  char *line = strtok(header_buf, "\r\n");
  if (!line || sscanf(line, "%15s %255s", req.method, req.path) != 2) { req.error_status = 400; return req; }

  size_t content_length = 0;
  while ((line = strtok(NULL, "\r\n")) != NULL) {
    char *colon = strchr(line, ':');
    if (!colon) continue;
    *colon = '\0';
    char key[128];
    snprintf(key, sizeof(key), "%s", line);
    lower_ascii(key);
    char *value = colon + 1;
    trim(value);
    if (strcmp(key, "content-length") == 0) content_length = (size_t)strtoull(value, NULL, 10);
    else if (strcmp(key, "content-type") == 0) set_header_value(req.content_type, sizeof(req.content_type), value);
    else if (strcmp(key, "user-agent") == 0) set_header_value(req.user_agent, sizeof(req.user_agent), value);
  }

  if (content_length > max_body_bytes) { req.error_status = 413; return req; }
  req.body = (char *)calloc(content_length + 1, 1);
  if (!req.body && content_length > 0) { req.error_status = 500; return req; }

  size_t already = total - (header_end + 4);
  if (already > content_length) already = content_length;
  if (already > 0) memcpy(req.body, header_buf + header_end + 4, already);
  req.body_len = already;

  while (req.body_len < content_length) {
    char chunk[CHUNK_SIZE];
    n = recv(fd, chunk, sizeof(chunk), 0);
    if (n <= 0) { req.error_status = 400; free_request(&req); return req; }
    size_t to_copy = (size_t)n;
    if (req.body_len + to_copy > content_length) to_copy = content_length - req.body_len;
    memcpy(req.body + req.body_len, chunk, to_copy);
    req.body_len += to_copy;
  }
  return req;
}

static int append_ndjson(const char *path, const struct request *req, const struct sockaddr_storage *peer) {
  if (!ensure_parent_dirs(path)) return 0;
  FILE *out = fopen(path, "ab");
  if (!out) return 0;

  char ts[32];
  char remote[INET6_ADDRSTRLEN];
  now_iso8601(ts, sizeof(ts));
  remote_addr_str(peer, remote, sizeof(remote));

  fputs("{\"ts\":\"", out); json_escape_file(out, ts, strlen(ts));
  fputs("\",\"remote_addr\":\"", out); json_escape_file(out, remote, strlen(remote));
  fputs("\",\"content_type\":\"", out); json_escape_file(out, req->content_type, strlen(req->content_type));
  fputs("\",\"user_agent\":\"", out); json_escape_file(out, req->user_agent, strlen(req->user_agent));
  fputs("\",\"path\":\"", out); json_escape_file(out, req->path, strlen(req->path));
  fputs("\",\"raw_body\":\"", out); json_escape_file(out, req->body ? req->body : "", req->body_len);
  fputs("\"}\n", out);

  int ok = (fclose(out) == 0);
  return ok;
}

int main(void) {
  signal(SIGPIPE, SIG_IGN);
  int port = atoi(getenv_or("PORT", "8080"));
  const char *log_path = getenv_or("LOG_PATH", "/data/frontend-errors.ndjson");
  size_t max_body_bytes = getenv_size("MAX_BODY_BYTES", 65536);

  int server_fd = socket(AF_INET, SOCK_STREAM, 0);
  if (server_fd < 0) return 1;
  int yes = 1;
  setsockopt(server_fd, SOL_SOCKET, SO_REUSEADDR, &yes, sizeof(yes));

  struct sockaddr_in addr;
  memset(&addr, 0, sizeof(addr));
  addr.sin_family = AF_INET;
  addr.sin_addr.s_addr = htonl(INADDR_ANY);
  addr.sin_port = htons((uint16_t)port);
  if (bind(server_fd, (struct sockaddr *)&addr, sizeof(addr)) != 0) return 1;
  if (listen(server_fd, 128) != 0) return 1;

  fprintf(stderr, "js-error-collector listening on :%d log=%s\n", port, log_path);
  while (1) {
    struct sockaddr_storage peer;
    socklen_t peer_len = sizeof(peer);
    int client_fd = accept(server_fd, (struct sockaddr *)&peer, &peer_len);
    if (client_fd < 0) continue;

    struct request req = read_request(client_fd, max_body_bytes);
    if (req.error_status == 431) send_response(client_fd, 431, "Request Header Fields Too Large", "{\"error\":\"headers too large\"}", NULL);
    else if (req.error_status == 413) send_response(client_fd, 413, "Payload Too Large", "{\"error\":\"payload too large\"}", NULL);
    else if (req.error_status == 500) send_response(client_fd, 500, "Internal Server Error", "{\"error\":\"allocation failed\"}", NULL);
    else if (req.error_status != 0) send_response(client_fd, 400, "Bad Request", "{\"error\":\"bad request\"}", NULL);
    else if (strcmp(req.method, "GET") == 0 && (strcmp(req.path, "/health") == 0 || strcmp(req.path, "/ready") == 0)) {
      send_response(client_fd, 200, "OK", "{\"status\":\"ok\"}", NULL);
    } else if (strcmp(req.path, "/frontend-error") == 0 && strcmp(req.method, "POST") == 0) {
      if (append_ndjson(log_path, &req, &peer)) send_response(client_fd, 202, "Accepted", "{\"status\":\"received\"}", NULL);
      else send_response(client_fd, 500, "Internal Server Error", "{\"error\":\"log write failed\"}", NULL);
    } else if (strcmp(req.path, "/frontend-error") == 0) {
      send_response(client_fd, 405, "Method Not Allowed", "{\"error\":\"method not allowed\"}", NULL);
    } else {
      send_response(client_fd, 404, "Not Found", "{\"error\":\"not found\"}", NULL);
    }

    free_request(&req);
    close(client_fd);
  }
}