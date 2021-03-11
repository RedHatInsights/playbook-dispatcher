
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
    "event_type",
    "payload"
})
public class RunEvent {

    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("event_type")
    private RunEvent.EventType eventType;
    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("payload")
    private Payload payload;
    @JsonIgnore
    private Map<String, Object> additionalProperties = new HashMap<String, Object>();

    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("event_type")
    public RunEvent.EventType getEventType() {
        return eventType;
    }

    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("event_type")
    public void setEventType(RunEvent.EventType eventType) {
        this.eventType = eventType;
    }

    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("payload")
    public Payload getPayload() {
        return payload;
    }

    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("payload")
    public void setPayload(Payload payload) {
        this.payload = payload;
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
        sb.append(RunEvent.class.getName()).append('@').append(Integer.toHexString(System.identityHashCode(this))).append('[');
        sb.append("eventType");
        sb.append('=');
        sb.append(((this.eventType == null)?"<null>":this.eventType));
        sb.append(',');
        sb.append("payload");
        sb.append('=');
        sb.append(((this.payload == null)?"<null>":this.payload));
        sb.append(',');
        sb.append("additionalProperties");
        sb.append('=');
        sb.append(((this.additionalProperties == null)?"<null>":this.additionalProperties));
        sb.append(',');
        if (sb.charAt((sb.length()- 1)) == ',') {
            sb.setCharAt((sb.length()- 1), ']');
        } else {
            sb.append(']');
        }
        return sb.toString();
    }

    @Override
    public int hashCode() {
        int result = 1;
        result = ((result* 31)+((this.eventType == null)? 0 :this.eventType.hashCode()));
        result = ((result* 31)+((this.additionalProperties == null)? 0 :this.additionalProperties.hashCode()));
        result = ((result* 31)+((this.payload == null)? 0 :this.payload.hashCode()));
        return result;
    }

    @Override
    public boolean equals(Object other) {
        if (other == this) {
            return true;
        }
        if ((other instanceof RunEvent) == false) {
            return false;
        }
        RunEvent rhs = ((RunEvent) other);
        return ((((this.eventType == rhs.eventType)||((this.eventType!= null)&&this.eventType.equals(rhs.eventType)))&&((this.additionalProperties == rhs.additionalProperties)||((this.additionalProperties!= null)&&this.additionalProperties.equals(rhs.additionalProperties))))&&((this.payload == rhs.payload)||((this.payload!= null)&&this.payload.equals(rhs.payload))));
    }

    public enum EventType {

        CREATE("create"),
        READ("read"),
        UPDATE("update"),
        DELETE("delete");
        private final String value;
        private final static Map<String, RunEvent.EventType> CONSTANTS = new HashMap<String, RunEvent.EventType>();

        static {
            for (RunEvent.EventType c: values()) {
                CONSTANTS.put(c.value, c);
            }
        }

        private EventType(String value) {
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
        public static RunEvent.EventType fromValue(String value) {
            RunEvent.EventType constant = CONSTANTS.get(value);
            if (constant == null) {
                throw new IllegalArgumentException(value);
            } else {
                return constant;
            }
        }

    }

}
