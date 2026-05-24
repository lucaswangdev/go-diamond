import java.io.*;
import java.net.HttpURLConnection;
import java.net.URL;
import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;
import java.util.*;
import java.util.function.Consumer;

public class DiamondExample {

    private final String serverUrl;
    private final String namespace;
    private final String group;
    private final String dataId;
    private final Map<String, String> cache = new HashMap<>();
    private final Map<String, String> md5Cache = new HashMap<>();
    private final List<Consumer<String>> listeners = new ArrayList<>();
    private volatile boolean running = false;
    private Thread watchThread;

    public DiamondExample(String serverUrl, String namespace, String group, String dataId) {
        this.serverUrl = serverUrl.endsWith("/") ? serverUrl.substring(0, serverUrl.length() - 1) : serverUrl;
        this.namespace = namespace;
        this.group = group;
        this.dataId = dataId;
    }

    public String getConfig() throws IOException {
        String urlStr = String.format("%s/api/v1/configs/%s/%s/%s",
            serverUrl, namespace, group, dataId);

        URL url = new URL(urlStr);
        HttpURLConnection conn = (HttpURLConnection) url.openConnection();
        conn.setRequestMethod("GET");
        conn.setConnectTimeout(30000);
        conn.setReadTimeout(30000);

        try {
            int responseCode = conn.getResponseCode();
            if (responseCode == 200) {
                String response = readResponse(conn);
                Map<String, Object> json = parseJson(response);
                if (json != null && "0".equals(String.valueOf(json.get("code")))) {
                    Map<String, Object> data = (Map<String, Object>) json.get("data");
                    if (data != null) {
                        String content = (String) data.get("content");
                        String md5 = (String) data.get("contentMd5");
                        String key = namespace + "/" + group + "/" + dataId;
                        cache.put(key, content);
                        md5Cache.put(key, md5);
                        return content;
                    }
                }
            } else if (responseCode == 404) {
                return null;
            }
        } finally {
            conn.disconnect();
        }
        return null;
    }

    public void addListener(Consumer<String> listener) {
        listeners.add(listener);
    }

    public void startWatch() {
        if (running) return;
        running = true;
        watchThread = new Thread(this::watchLoop);
        watchThread.setDaemon(true);
        watchThread.start();
    }

    public void stopWatch() {
        running = false;
        if (watchThread != null) {
            watchThread.interrupt();
            try {
                watchThread.join(5000);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
        }
    }

    private void watchLoop() {
        while (running) {
            try {
                String config = getConfig();
                if (config != null) {
                    String key = namespace + "/" + group + "/" + dataId;
                    String oldMd5 = md5Cache.get(key);
                    String newMd5 = md5(config);

                    if (oldMd5 != null && !oldMd5.equals(newMd5)) {
                        notifyListeners(config);
                    }
                }
                Thread.sleep(2000);
            } catch (InterruptedException e) {
                break;
            } catch (Exception e) {
                System.err.println("Watch error: " + e.getMessage());
            }
        }
    }

    private void notifyListeners(String config) {
        for (Consumer<String> listener : listeners) {
            try {
                listener.accept(config);
            } catch (Exception e) {
                System.err.println("Listener error: " + e.getMessage());
            }
        }
    }

    public String getCached() {
        String key = namespace + "/" + group + "/" + dataId;
        return cache.get(key);
    }

    public String watchWithMd5(int timeout) throws IOException {
        String key = namespace + "/" + group + "/" + dataId;
        String currentMd5 = md5Cache.getOrDefault(key, "");

        String urlStr = String.format("%s/api/v1/watch/%s/%s/%s?md5=%s&timeout=%d",
            serverUrl, namespace, group, dataId,
            URLEncoder.encode(currentMd5, "UTF-8"), timeout);

        URL url = new URL(urlStr);
        HttpURLConnection conn = (HttpURLConnection) url.openConnection();
        conn.setRequestMethod("GET");
        conn.setConnectTimeout((timeout + 5) * 1000);
        conn.setReadTimeout((timeout + 5) * 1000);

        try {
            int responseCode = conn.getResponseCode();
            if (responseCode == 200) {
                String response = readResponse(conn);
                Map<String, Object> json = parseJson(response);
                if (json != null && "0".equals(String.valueOf(json.get("code")))) {
                    Map<String, Object> data = (Map<String, Object>) json.get("data");
                    return data != null ? (String) data.get("content") : null;
                }
            } else if (responseCode == 304) {
                return null; // Not modified
            }
        } finally {
            conn.disconnect();
        }
        return null;
    }

    public String batchWatch(List<Map<String, String>> items, int timeout) throws IOException {
        String urlStr = serverUrl + "/api/v1/watch/batch";
        URL url = new URL(urlStr);
        HttpURLConnection conn = (HttpURLConnection) url.openConnection();
        conn.setRequestMethod("POST");
        conn.setRequestProperty("Content-Type", "application/json");
        conn.setConnectTimeout((timeout + 5) * 1000);
        conn.setReadTimeout((timeout + 5) * 1000);
        conn.setDoOutput(true);

        StringBuilder jsonBuilder = new StringBuilder();
        jsonBuilder.append("{\"watchItems\":[");
        for (int i = 0; i < items.size(); i++) {
            Map<String, String> item = items.get(i);
            if (i > 0) jsonBuilder.append(",");
            jsonBuilder.append("{");
            jsonBuilder.append("\"namespace\":\"").append(item.get("namespace")).append("\",");
            jsonBuilder.append("\"group\":\"").append(item.get("group")).append("\",");
            jsonBuilder.append("\"dataId\":\"").append(item.get("dataId")).append("\",");
            jsonBuilder.append("\"md5\":\"").append(item.getOrDefault("md5", "")).append("\"");
            jsonBuilder.append("}");
        }
        jsonBuilder.append("],\"timeout\":").append(timeout).append("}");

        try (OutputStream os = conn.getOutputStream()) {
            os.write(jsonBuilder.toString().getBytes(StandardCharsets.UTF_8));
        }

        try {
            int responseCode = conn.getResponseCode();
            if (responseCode == 200) {
                String response = readResponse(conn);
                return response;
            }
        } finally {
            conn.disconnect();
        }
        return null;
    }

    private static String readResponse(HttpURLConnection conn) throws IOException {
        StringBuilder response = new StringBuilder();
        try (BufferedReader br = new BufferedReader(
                new InputStreamReader(conn.getInputStream(), StandardCharsets.UTF_8))) {
            String line;
            while ((line = br.readLine()) != null) {
                response.append(line);
            }
        }
        return response.toString();
    }

    private static Map<String, Object> parseJson(String jsonStr) {
        Map<String, Object> result = new HashMap<>();
        // Simple JSON parser for demo purposes
        // In production, use Jackson or Gson
        try {
            // Handle {"code":0,"data":{...}}
            int codeIdx = jsonStr.indexOf("\"code\":");
            int dataIdx = jsonStr.indexOf("\"data\":");

            if (codeIdx >= 0) {
                int colonIdx = jsonStr.indexOf(":", codeIdx);
                int commaIdx = jsonStr.indexOf(",", colonIdx);
                if (commaIdx < 0) commaIdx = jsonStr.indexOf("}", colonIdx);
                String code = jsonStr.substring(colonIdx + 1, commaIdx).trim();
                result.put("code", code);
            }

            if (dataIdx >= 0) {
                int braceStart = jsonStr.indexOf("{", dataIdx);
                int braceEnd = findMatchingBrace(jsonStr, braceStart);
                if (braceStart >= 0 && braceEnd > braceStart) {
                    String dataStr = jsonStr.substring(braceStart, braceEnd + 1);
                    result.put("data_content", dataStr);
                }
            }
        } catch (Exception e) {
            System.err.println("JSON parse error: " + e.getMessage());
        }
        return result;
    }

    private static int findMatchingBrace(String str, int start) {
        int depth = 1;
        for (int i = start + 1; i < str.length(); i++) {
            char c = str.charAt(i);
            if (c == '{') depth++;
            else if (c == '}') {
                depth--;
                if (depth == 0) return i;
            }
        }
        return -1;
    }

    private static String md5(String input) {
        try {
            java.security.MessageDigest md = java.security.MessageDigest.getInstance("MD5");
            byte[] digest = md.digest(input.getBytes(StandardCharsets.UTF_8));
            StringBuilder sb = new StringBuilder();
            for (byte b : digest) {
                sb.append(String.format("%02x", b));
            }
            return sb.toString();
        } catch (Exception e) {
            return "";
        }
    }

    // Demo methods
    public static void demoBasicUsage() {
        printSeparator();
        System.out.println("Demo: Basic Configuration Retrieval");
        printSeparator();

        try {
            DiamondExample client = new DiamondExample(
                "http://127.0.0.1:8080", "default", "DEFAULT_GROUP", "app.json");

            String config = client.getConfig();
            if (config != null) {
                System.out.println("Config content: " + config);
            } else {
                System.out.println("Config not found or server unavailable");
            }
        } catch (Exception e) {
            System.err.println("Error: " + e.getMessage());
        }
    }

    public static void demoWithCache() {
        printSeparator();
        System.out.println("Demo: Configuration with Cache");
        printSeparator();

        try {
            DiamondExample client = new DiamondExample(
                "http://127.0.0.1:8080", "default", "DEFAULT_GROUP", "app.json");

            String config = client.getConfig();
            System.out.println("First fetch: " + config);

            String cached = client.getCached();
            System.out.println("Cached value: " + cached);
        } catch (Exception e) {
            System.err.println("Error: " + e.getMessage());
        }
    }

    public static void demoListener() {
        printSeparator();
        System.out.println("Demo: Configuration Change Listener");
        printSeparator();

        try {
            DiamondExample client = new DiamondExample(
                "http://127.0.0.1:8080", "default", "DEFAULT_GROUP", "app.json");

            client.addListener(newContent -> {
                System.out.println("Config changed! New content: " + newContent);
            });

            String config = client.getConfig();
            System.out.println("Initial config: " + config);

            client.startWatch();
            System.out.println("Watching for changes... (10 seconds)");

            Thread.sleep(10000);
            client.stopWatch();
            System.out.println("Stopped watching");
        } catch (Exception e) {
            System.err.println("Error: " + e.getMessage());
        }
    }

    public static void demoBatchWatch() {
        printSeparator();
        System.out.println("Demo: Batch Watching Multiple Configs");
        printSeparator();

        try {
            DiamondExample client = new DiamondExample(
                "http://127.0.0.1:8080", "default", "DEFAULT_GROUP", "app.json");

            List<Map<String, String>> items = new ArrayList<>();
            items.add(createItem("default", "DEFAULT_GROUP", "app.json"));
            items.add(createItem("default", "DEFAULT_GROUP", "db.json"));
            items.add(createItem("default", "DEFAULT_GROUP", "redis.json"));

            String response = client.batchWatch(items, 30);
            System.out.println("Batch watch response: " + response);
        } catch (Exception e) {
            System.err.println("Error: " + e.getMessage());
        }
    }

    private static String repeat(String str, int count) {
        StringBuilder sb = new StringBuilder();
        for (int i = 0; i < count; i++) {
            sb.append(str);
        }
        return sb.toString();
    }

    private static void printSeparator() {
        System.out.println("\n" + repeat("=", 50));
    }

    private static Map<String, String> createItem(String ns, String group, String dataId) {
        Map<String, String> item = new HashMap<>();
        item.put("namespace", ns);
        item.put("group", group);
        item.put("dataId", dataId);
        item.put("md5", "");
        return item;
    }

    public static void main(String[] args) {
        System.out.println("go-diamond Java Client Demo");
        printSeparator();
        System.out.println();
        System.out.println("Note: Make sure go-diamond server is running on http://127.0.0.1:8080");
        System.out.println();

        demoBasicUsage();
        demoWithCache();
        demoBatchWatch();
        demoListener();
    }
}