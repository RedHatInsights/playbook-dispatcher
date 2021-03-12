package com.redhat.cloud.platform.playbook_dispatcher.config;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.Map;
import java.util.Set;
import java.util.stream.Collectors;

import org.apache.kafka.common.config.ConfigData;
import org.apache.kafka.common.config.ConfigException;
import org.apache.kafka.common.config.provider.ConfigProvider;

/**
 * {@link ConfigProvider} implementation that reads values from files.
 * The entire content of a file is used as the value for the given key.
 *
 * Unlike {@link org.apache.kafka.common.config.provider.FileConfigProvider} this implementation does not require
 * the files to be a .properties file and is therefore suitable for reading Kubernetes opaque secrets that are
 * mounted to the filesystem.
 *
 * Example:
 *
 * "database.password": "${file:/opt/kafka/external-configuration/db/db.password}",
 */
public class PlainFileConfigProvider implements ConfigProvider {

    @Override
    public void configure(Map<String, ?> cfg) {
    }

    @Override
    public void close() throws IOException {
    }

    @Override
    public ConfigData get(String path) {
        throw new UnsupportedOperationException();
    }

    @Override
    @SuppressWarnings("PMD.PreserveStackTrace")
    public ConfigData get(String path, Set<String> keys) {
        if (path == null || path.length() != 0) {
            throw new IllegalArgumentException("Only basic syntax (e.g. ${file:/path/to/file}) is supported");
        }

        return new ConfigData(keys.stream().collect(Collectors.toMap(key -> key, key -> {
            try {
                return new String(Files.readAllBytes(Paths.get(key))).trim();
            } catch (IOException e) {
                throw new ConfigException("Error reading " + key + " due to " + e.toString());
            }
        })));
    }
}
