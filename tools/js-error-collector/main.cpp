#include <arpa/inet.h>
#include <cerrno>
#include <cctype>
#include <chrono>
#include <csignal>
#include <cstdio>
#include <cstdlib>
#include <ctime>
#include <fstream>
#include <iostream>
#include <netinet/in.h>
#include <sstream>
#include <string>
#include <sys/socket.h>
#include <sys/stat.h>
#include <unistd.h>
#include <unordered_map>

namespace {

struct Request {
  std::string method;
  std::string path;
  std::string body;
  std::string content_type;
  std::string user_agent;
  int error_status = 0;
};

std::string getenv_or(const char* key, const std::string& fallback) {
  const char* value = std::getenv(key);
  return (value && *value) ? value : fallback;
}

std::size_t getenv_size(const char* key, std::size_t fallback) {
  const char* value = std::getenv(key);
  if (!value || !*value) return fallback;
  try { return static_cast<std::size_t>(std::stoull(value)); } catch (...) { return fallback; }
}

std::string trim(const std::string& s) {
  auto start = s.find_first_not_of(" \t\r\n");
  if (start == std::string::npos) return "";
  auto end = s.find_last_not_of(" \t\r\n");
  return s.substr(start, end - start + 1);
}

std::string json_escape(const std::string& input) {
  std::ostringstream out;
  for (unsigned char c : input) {
    switch (c) {
      case '"': out << "\\\""; break;
      case '\\': out << "\\\\"; break;
      case '\b': out << "\\b"; break;
      case '\f': out << "\\f"; break;
      case '\n': out << "\\n"; break;
      case '\r': out << "\\r"; break;
      case '\t': out << "\\t"; break;
      default:
        if (c < 0x20) {
          char buf[7];
          std::snprintf(buf, sizeof(buf), "\\u%04x", c);
          out << buf;
        } else {
          out << static_cast<char>(c);
        }
    }
  }
  return out.str();
}

std::string now_utc_iso8601() {
  auto now = std::chrono::system_clock::now();
  std::time_t t = std::chrono::system_clock::to_time_t(now);
  std::tm tm{};
  gmtime_r(&t, &tm);
  char buf[32];
  std::strftime(buf, sizeof(buf), "%Y-%m-%dT%H:%M:%SZ", &tm);
  return buf;
}

bool ensure_parent_dirs(const std::string& path) {
  std::size_t pos = 0;
  while ((pos = path.find('/', pos + 1)) != std::string::npos) {
    auto part = path.substr(0, pos);
    if (part.empty()) continue;
    if (::mkdir(part.c_str(), 0755) != 0 && errno != EEXIST) return false;
  }
  return true;
}

void send_response(int fd, int status, const std::string& reason, const std::string& body,
                   const std::string& content_type = "application/json") {
  std::ostringstream resp;
  resp << "HTTP/1.1 " << status << ' ' << reason << "\r\n"
       << "Content-Type: " << content_type << "\r\n"
       << "Content-Length: " << body.size() << "\r\n"
       << "Connection: close\r\n\r\n"
       << body;
  auto data = resp.str();
  ::send(fd, data.data(), data.size(), 0);
}

std::string remote_addr(const sockaddr_storage& addr) {
  char buf[INET6_ADDRSTRLEN] = {0};
  if (addr.ss_family == AF_INET) {
    auto* a = reinterpret_cast<const sockaddr_in*>(&addr);
    inet_ntop(AF_INET, &a->sin_addr, buf, sizeof(buf));
  } else if (addr.ss_family == AF_INET6) {
    auto* a = reinterpret_cast<const sockaddr_in6*>(&addr);
    inet_ntop(AF_INET6, &a->sin6_addr, buf, sizeof(buf));
  }
  return buf[0] ? std::string(buf) : "unknown";
}

Request read_request(int fd, std::size_t max_body_bytes) {
  Request req;
  std::string data;
  char chunk[4096];
  std::size_t header_end = std::string::npos;

  while ((header_end = data.find("\r\n\r\n")) == std::string::npos) {
    auto n = ::recv(fd, chunk, sizeof(chunk), 0);
    if (n <= 0) { req.error_status = 400; return req; }
    data.append(chunk, static_cast<std::size_t>(n));
    if (data.size() > 16 * 1024) { req.error_status = 431; return req; }
  }

  std::string head = data.substr(0, header_end);
  std::istringstream lines(head);
  std::string line;
  if (!std::getline(lines, line)) { req.error_status = 400; return req; }
  if (!line.empty() && line.back() == '\r') line.pop_back();
  std::istringstream first(line);
  if (!(first >> req.method >> req.path)) { req.error_status = 400; return req; }

  std::unordered_map<std::string, std::string> headers;
  while (std::getline(lines, line)) {
    if (!line.empty() && line.back() == '\r') line.pop_back();
    auto colon = line.find(':');
    if (colon == std::string::npos) continue;
    auto key = line.substr(0, colon);
    for (char& c : key) c = static_cast<char>(std::tolower(static_cast<unsigned char>(c)));
    headers[key] = trim(line.substr(colon + 1));
  }

  req.content_type = headers["content-type"];
  req.user_agent = headers["user-agent"];
  std::size_t content_length = 0;
  if (headers.count("content-length")) {
    try { content_length = static_cast<std::size_t>(std::stoull(headers["content-length"])); }
    catch (...) { req.error_status = 400; return req; }
  }
  if (content_length > max_body_bytes) { req.error_status = 413; return req; }

  req.body = data.substr(header_end + 4);
  while (req.body.size() < content_length) {
    auto n = ::recv(fd, chunk, sizeof(chunk), 0);
    if (n <= 0) { req.error_status = 400; return req; }
    req.body.append(chunk, static_cast<std::size_t>(n));
  }
  if (req.body.size() > content_length) req.body.resize(content_length);
  return req;
}

bool append_ndjson(const std::string& path, const std::string& line) {
  if (!ensure_parent_dirs(path)) return false;
  std::ofstream out(path, std::ios::app | std::ios::binary);
  if (!out) return false;
  out << line << '\n';
  return out.good();
}

}  // namespace

int main() {
  std::signal(SIGPIPE, SIG_IGN);
  const int port = std::stoi(getenv_or("PORT", "8080"));
  const std::string log_path = getenv_or("LOG_PATH", "/data/frontend-errors.ndjson");
  const std::size_t max_body_bytes = getenv_size("MAX_BODY_BYTES", 65536);

  int server_fd = ::socket(AF_INET, SOCK_STREAM, 0);
  if (server_fd < 0) return 1;
  int yes = 1;
  setsockopt(server_fd, SOL_SOCKET, SO_REUSEADDR, &yes, sizeof(yes));

  sockaddr_in addr{};
  addr.sin_family = AF_INET;
  addr.sin_addr.s_addr = htonl(INADDR_ANY);
  addr.sin_port = htons(static_cast<uint16_t>(port));
  if (::bind(server_fd, reinterpret_cast<sockaddr*>(&addr), sizeof(addr)) != 0) return 1;
  if (::listen(server_fd, 128) != 0) return 1;

  std::cerr << "js-error-collector listening on :" << port << " log=" << log_path << std::endl;
  while (true) {
    sockaddr_storage peer{};
    socklen_t peer_len = sizeof(peer);
    int client_fd = ::accept(server_fd, reinterpret_cast<sockaddr*>(&peer), &peer_len);
    if (client_fd < 0) continue;

    Request req = read_request(client_fd, max_body_bytes);
    if (req.error_status == 431) {
      send_response(client_fd, 431, "Request Header Fields Too Large", "{\"error\":\"headers too large\"}");
    } else if (req.error_status == 413) {
      send_response(client_fd, 413, "Payload Too Large", "{\"error\":\"payload too large\"}");
    } else if (req.error_status != 0) {
      send_response(client_fd, 400, "Bad Request", "{\"error\":\"bad request\"}");
    } else if (req.method == "GET" && (req.path == "/health" || req.path == "/ready")) {
      send_response(client_fd, 200, "OK", "{\"status\":\"ok\"}");
    } else if (req.path == "/frontend-error" && req.method == "POST") {
      std::ostringstream line;
      line << "{"
           << "\"ts\":\"" << json_escape(now_utc_iso8601()) << "\","
           << "\"remote_addr\":\"" << json_escape(remote_addr(peer)) << "\","
           << "\"content_type\":\"" << json_escape(req.content_type) << "\","
           << "\"user_agent\":\"" << json_escape(req.user_agent) << "\","
           << "\"path\":\"" << json_escape(req.path) << "\","
           << "\"raw_body\":\"" << json_escape(req.body) << "\"}";
      std::string entry = line.str();
      if (append_ndjson(log_path, entry)) {
        send_response(client_fd, 202, "Accepted", "{\"status\":\"received\"}");
      } else {
        send_response(client_fd, 500, "Internal Server Error", "{\"error\":\"log write failed\"}");
      }
    } else if (req.path == "/frontend-error") {
      send_response(client_fd, 405, "Method Not Allowed", "{\"error\":\"method not allowed\"}");
    } else {
      send_response(client_fd, 404, "Not Found", "{\"error\":\"not found\"}");
    }
    ::close(client_fd);
  }
}