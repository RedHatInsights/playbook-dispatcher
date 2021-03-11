package com.redhat.cloud.platform.playbook_dispatcher;

import java.util.HashMap;
import java.util.Map;

import org.apache.kafka.connect.data.Schema;
import org.apache.kafka.connect.data.SchemaBuilder;
import org.apache.kafka.connect.data.Struct;

class StructBuilder {

    private final Map<String, Object> values = new HashMap<>();
    private final SchemaBuilder schemaBuilder = SchemaBuilder.struct();

    public StructBuilder put(String key, Object value) {
        schemaBuilder.field(key, schemaForValue(value));
        values.put(key, value);
        return this;
    }

    private Schema schemaForValue(Object value) {
        switch (value.getClass().getName()) {
            case "java.lang.String":
                return SchemaBuilder.string();
            case "java.lang.Integer":
                return SchemaBuilder.int32();
            case "org.apache.kafka.connect.data.Struct":
                return ((Struct) value).schema();
            default:
                throw new IllegalArgumentException("Unsupported value type: " + value.getClass().getName());
        }
    }

    public Struct build() {
        Struct result = new Struct(schemaBuilder.build());

        for (Map.Entry<String, Object> entry : values.entrySet()) {
            result.put(entry.getKey(), entry.getValue());
        }

        return result;
    }
}
