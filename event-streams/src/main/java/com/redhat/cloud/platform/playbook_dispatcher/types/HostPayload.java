package com.redhat.cloud.platform.playbook_dispatcher.types;

import java.util.HashMap;
import java.util.Map;
import com.fasterxml.jackson.annotation.JsonAnyGetter;
import com.fasterxml.jackson.annotation.JsonAnySetter;
import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonPropertyOrder;
import com.fasterxml.jackson.annotation.JsonValue;

@JsonInclude(JsonInclude.Include.NON_NULL)
@JsonPropertyOrder({
        "id",
        "run_id",
        "inventory_id",
        "host",
        "log",
        "events",
        "sat_sequence",
        "status",
        "timeout",
        "created_at",
        "updated_at"
})
public class HostPayload {

    /**
     *
     * (Required)
     *
     */
    @JsonProperty("id")
    private String id;
    @JsonProperty("run_id")
    private String runId;
    /**
     *
     * (Required)
     *
     */
    @JsonProperty("inventory_id")
    private String inventoryId;
    /**
     *
     * (Required)
     *
     */
    @JsonProperty("host")
    private String host;
    @JsonProperty("log")
    private String log;
    /**
     *
     * (Required)
     *
     */
    @JsonProperty("events")
    private String events;
    @JsonProperty("sat_sequence")
    private Object satSequence;
    /**
     *
     * (Required)
     * (Required)
     *
     */
    @JsonProperty("status")
    private HostPayload.Status status;
    @JsonProperty("timeout")
    private Integer timeout;
    /**
     *
     * (Required)
     *
     */
    @JsonProperty("created_at")
    private String createdAt;
    /**
     *
     * (Required)
     *
     */
    @JsonProperty("updated_at")
    private String updatedAt;
    @JsonIgnore
    private Map<String, Object> additionalProperties = new HashMap<String, Object>();

    /**
     *
     * (Required)
     *
     */
    @JsonProperty("id")
    public String getId() {
        return id;
    }

    /**
     *
     * (Required)
     *
     */
    @JsonProperty("id")
    public void setId(String id) {
        this.id = id;
    }

    @JsonProperty("run_id")
    public String getRunId() {
        return runId;
    }

    @JsonProperty("run_id")
    public void setRunId(String runId) {
        this.runId = runId;
    }

    /**
     *
     * (Required)
     *
     */
    @JsonProperty("inventory_id")
    public String getInventoryId() {
        return inventoryId;
    }

    /**
     *
     * (Required)
     *
     */
    @JsonProperty("inventory_id")
    public void setInventoryId(String inventoryId) {
        this.inventoryId = inventoryId;
    }

    /**
     *
     * (Required)
     *
     */
    @JsonProperty("host")
    public String getHost() {
        return host;
    }

    /**
     *
     * (Required)
     *
     */
    @JsonProperty("host")
    public void setHost(String host) {
        this.host = host;
    }

    @JsonProperty("log")
    public String getLog() {
        return log;
    }

    @JsonProperty("log")
    public void setLog(String log) {
        this.log = log;
    }

    /**
     *
     * (Required)
     *
     */
    @JsonProperty("events")
    public String getEvents() {
        return events;
    }

    /**
     *
     * (Required)
     *
     */
    @JsonProperty("events")
    public void setEvents(String events) {
        this.events = events;
    }

    @JsonProperty("sat_sequence")
    public Object getSatSequence() {
        return satSequence;
    }

    @JsonProperty("sat_sequence")
    public void setSatSequence(Object satSequence) {
        this.satSequence = satSequence;
    }

    /**
     *
     * (Required)
     *
     */
    @JsonProperty("status")
    public HostPayload.Status getStatus() {
        return status;
    }

    /**
     *
     * (Required)
     *
     */
    @JsonProperty("status")
    public void setStatus(HostPayload.Status status) {
        this.status = status;
    }

    @JsonProperty("timeout")
    public Integer getTimeout() {
        return timeout;
    }

    @JsonProperty("timeout")
    public void setTimeout(Integer timeout) {
        this.timeout = timeout;
    }

    /**
     *
     * (Required)
     *
     */
    @JsonProperty("created_at")
    public String getCreatedAt() {
        return createdAt;
    }

    /**
     *
     * (Required)
     *
     */
    @JsonProperty("created_at")
    public void setCreatedAt(String createdAt) {
        this.createdAt = createdAt;
    }

    /**
     *
     * (Required)
     *
     */
    @JsonProperty("updated_at")
    public String getUpdatedAt() {
        return updatedAt;
    }

    /**
     *
     * (Required)
     *
     */
    @JsonProperty("updated_at")
    public void setUpdatedAt(String updatedAt) {
        this.updatedAt = updatedAt;
    }

    @JsonAnyGetter
    public Map<String, Object> getAdditionalProperties() {
        return this.additionalProperties;
    }

    @JsonAnySetter
    public void setAdditionalProperty(String name, Object value) {
        this.additionalProperties.put(name, value);
    }

    @Override
    public String toString() {
        StringBuilder sb = new StringBuilder();
        sb.append(HostPayload.class.getName()).append('@').append(Integer.toHexString(System.identityHashCode(this)))
                .append('[');
        sb.append("id");
        sb.append('=');
        sb.append(((this.id == null) ? "<null>" : this.id));
        sb.append(',');
        sb.append("runId");
        sb.append('=');
        sb.append(((this.runId == null) ? "<null>" : this.runId));
        sb.append(',');
        sb.append("inventoryId");
        sb.append('=');
        sb.append(((this.inventoryId == null) ? "<null>" : this.inventoryId));
        sb.append(',');
        sb.append("host");
        sb.append('=');
        sb.append(((this.host == null) ? "<null>" : this.host));
        sb.append(',');
        sb.append("log");
        sb.append('=');
        sb.append(((this.log == null) ? "<null>" : this.log));
        sb.append(',');
        sb.append("events");
        sb.append('=');
        sb.append(((this.events == null) ? "<null>" : this.events));
        sb.append(',');
        sb.append("satSequence");
        sb.append('=');
        sb.append(((this.satSequence == null) ? "<null>" : this.satSequence));
        sb.append(',');
        sb.append("status");
        sb.append('=');
        sb.append(((this.status == null) ? "<null>" : this.status));
        sb.append(',');
        sb.append("timeout");
        sb.append('=');
        sb.append(((this.timeout == null) ? "<null>" : this.timeout));
        sb.append(',');
        sb.append("createdAt");
        sb.append('=');
        sb.append(((this.createdAt == null) ? "<null>" : this.createdAt));
        sb.append(',');
        sb.append("updatedAt");
        sb.append('=');
        sb.append(((this.updatedAt == null) ? "<null>" : this.updatedAt));
        sb.append(',');
        sb.append("additionalProperties");
        sb.append('=');
        sb.append(((this.additionalProperties == null) ? "<null>" : this.additionalProperties));
        sb.append(',');
        if (sb.charAt((sb.length() - 1)) == ',') {
            sb.setCharAt((sb.length() - 1), ']');
        } else {
            sb.append(']');
        }
        return sb.toString();
    }

    @Override
    public int hashCode() {
        int result = 1;
        result = ((result * 31) + ((this.log == null) ? 0 : this.log.hashCode()));
        result = ((result * 31) + ((this.timeout == null) ? 0 : this.timeout.hashCode()));
        result = ((result * 31) + ((this.createdAt == null) ? 0 : this.createdAt.hashCode()));
        result = ((result * 31) + ((this.inventoryId == null) ? 0 : this.inventoryId.hashCode()));
        result = ((result * 31) + ((this.host == null) ? 0 : this.host.hashCode()));
        result = ((result * 31) + ((this.satSequence == null) ? 0 : this.satSequence.hashCode()));
        result = ((result * 31) + ((this.id == null) ? 0 : this.id.hashCode()));
        result = ((result * 31) + ((this.runId == null) ? 0 : this.runId.hashCode()));
        result = ((result * 31) + ((this.additionalProperties == null) ? 0 : this.additionalProperties.hashCode()));
        result = ((result * 31) + ((this.events == null) ? 0 : this.events.hashCode()));
        result = ((result * 31) + ((this.status == null) ? 0 : this.status.hashCode()));
        result = ((result * 31) + ((this.updatedAt == null) ? 0 : this.updatedAt.hashCode()));
        return result;
    }

    @Override
    public boolean equals(Object other) {
        if (other == this) {
            return true;
        }
        if ((other instanceof HostPayload) == false) {
            return false;
        }
        HostPayload rhs = ((HostPayload) other);
        return (((((((((((((this.log == rhs.log) || ((this.log != null) && this.log.equals(rhs.log)))
                && ((this.timeout == rhs.timeout) || ((this.timeout != null) && this.timeout.equals(rhs.timeout))))
                && ((this.createdAt == rhs.createdAt)
                        || ((this.createdAt != null) && this.createdAt.equals(rhs.createdAt))))
                && ((this.inventoryId == rhs.inventoryId)
                        || ((this.inventoryId != null) && this.inventoryId.equals(rhs.inventoryId))))
                && ((this.host == rhs.host) || ((this.host != null) && this.host.equals(rhs.host))))
                && ((this.satSequence == rhs.satSequence)
                        || ((this.satSequence != null) && this.satSequence.equals(rhs.satSequence))))
                && ((this.id == rhs.id) || ((this.id != null) && this.id.equals(rhs.id))))
                && ((this.runId == rhs.runId) || ((this.runId != null) && this.runId.equals(rhs.runId))))
                && ((this.additionalProperties == rhs.additionalProperties) || ((this.additionalProperties != null)
                        && this.additionalProperties.equals(rhs.additionalProperties))))
                && ((this.events == rhs.events) || ((this.events != null) && this.events.equals(rhs.events))))
                && ((this.status == rhs.status) || ((this.status != null) && this.status.equals(rhs.status))))
                && ((this.updatedAt == rhs.updatedAt)
                        || ((this.updatedAt != null) && this.updatedAt.equals(rhs.updatedAt))));
    }

    public enum Status {

        RUNNING("running"),
        SUCCESS("success"),
        FAILURE("failure"),
        TIMEOUT("timeout"),
        CANCELED("canceled");

        private final String value;
        private final static Map<String, HostPayload.Status> CONSTANTS = new HashMap<String, HostPayload.Status>();

        static {
            for (HostPayload.Status c : values()) {
                CONSTANTS.put(c.value, c);
            }
        }

        Status(String value) {
            this.value = value;
        }

        @Override
        public String toString() {
            return this.value;
        }

        @JsonValue
        public String value() {
            return this.value;
        }

        @JsonCreator
        public static HostPayload.Status fromValue(String value) {
            HostPayload.Status constant = CONSTANTS.get(value);
            if (constant == null) {
                throw new IllegalArgumentException(value);
            } else {
                return constant;
            }
        }

    }

}