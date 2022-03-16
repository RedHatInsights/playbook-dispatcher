
package com.redhat.cloud.platform.playbook_dispatcher.types;

import java.net.URI;
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
    "account",
    "recipient",
    "correlation_id",
    "service",
    "url",
    "labels",
    "playbook_name",
    "playbook_run_url",
    "sat_id",
    "sat_org_id",
    "status",
    "timeout",
    "created_at",
    "updated_at"
})
public class Payload {

    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("id")
    private String id;
    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("account")
    private String account;
    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("recipient")
    private String recipient;
    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("correlation_id")
    private String correlationId;
    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("service")
    private String service;
    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("url")
    private URI url;
    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("labels")
    private Labels labels;
    @JsonProperty("playbook_name")
    private String playbookName;
    @JsonProperty("playbook_run_url")
    private URI playbookRunUrl;
    @JsonProperty("sat_id")
    private String satId;
    @JsonProperty("sat_org_id")
    private String satOrgId;
    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("status")
    private Payload.Status status;
    /**
     * 
     * (Required)
     * 
     */
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

    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("account")
    public String getAccount() {
        return account;
    }

    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("account")
    public void setAccount(String account) {
        this.account = account;
    }

    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("recipient")
    public String getRecipient() {
        return recipient;
    }

    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("recipient")
    public void setRecipient(String recipient) {
        this.recipient = recipient;
    }

    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("correlation_id")
    public String getCorrelationId() {
        return correlationId;
    }

    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("correlation_id")
    public void setCorrelationId(String correlationId) {
        this.correlationId = correlationId;
    }

    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("service")
    public String getService() {
        return service;
    }

    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("service")
    public void setService(String service) {
        this.service = service;
    }

    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("url")
    public URI getUrl() {
        return url;
    }

    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("url")
    public void setUrl(URI url) {
        this.url = url;
    }

    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("labels")
    public Labels getLabels() {
        return labels;
    }

    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("labels")
    public void setLabels(Labels labels) {
        this.labels = labels;
    }

    @JsonProperty("playbook_name")
    public String getPlaybookName() {
        return playbookName;
    }

    @JsonProperty("playbook_name")
    public void setPlaybookName(String playbookName) {
        this.playbookName = playbookName;
    }

    @JsonProperty("playbook_run_url")
    public URI getPlaybookRunUrl() {
        return playbookRunUrl;
    }

    @JsonProperty("playbook_run_url")
    public void setPlaybookRunUrl(URI playbookRunUrl) {
        this.playbookRunUrl = playbookRunUrl;
    }

    @JsonProperty("sat_id")
    public String getSatId() {
        return satId;
    }

    @JsonProperty("sat_id")
    public void setSatId(String satId) {
        this.satId = satId;
    }

    @JsonProperty("sat_org_id")
    public String getSatOrgId() {
        return satOrgId;
    }

    @JsonProperty("sat_org_id")
    public void setSatOrgId(String satOrgId) {
        this.satOrgId = satOrgId;
    }

    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("status")
    public Payload.Status getStatus() {
        return status;
    }

    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("status")
    public void setStatus(Payload.Status status) {
        this.status = status;
    }

    /**
     * 
     * (Required)
     * 
     */
    @JsonProperty("timeout")
    public Integer getTimeout() {
        return timeout;
    }

    /**
     * 
     * (Required)
     * 
     */
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
        sb.append(Payload.class.getName()).append('@').append(Integer.toHexString(System.identityHashCode(this))).append('[');
        sb.append("id");
        sb.append('=');
        sb.append(((this.id == null)?"<null>":this.id));
        sb.append(',');
        sb.append("account");
        sb.append('=');
        sb.append(((this.account == null)?"<null>":this.account));
        sb.append(',');
        sb.append("recipient");
        sb.append('=');
        sb.append(((this.recipient == null)?"<null>":this.recipient));
        sb.append(',');
        sb.append("correlationId");
        sb.append('=');
        sb.append(((this.correlationId == null)?"<null>":this.correlationId));
        sb.append(',');
        sb.append("service");
        sb.append('=');
        sb.append(((this.service == null)?"<null>":this.service));
        sb.append(',');
        sb.append("url");
        sb.append('=');
        sb.append(((this.url == null)?"<null>":this.url));
        sb.append(',');
        sb.append("labels");
        sb.append('=');
        sb.append(((this.labels == null)?"<null>":this.labels));
        sb.append(',');
        sb.append("playbookName");
        sb.append('=');
        sb.append(((this.playbookName == null)?"<null>":this.playbookName));
        sb.append(',');
        sb.append("playbookRunUrl");
        sb.append('=');
        sb.append(((this.playbookRunUrl == null)?"<null>":this.playbookRunUrl));
        sb.append(',');
        sb.append("satId");
        sb.append('=');
        sb.append(((this.satId == null)?"<null>":this.satId));
        sb.append(',');
        sb.append("satOrgId");
        sb.append('=');
        sb.append(((this.satOrgId == null)?"<null>":this.satOrgId));
        sb.append(',');
        sb.append("status");
        sb.append('=');
        sb.append(((this.status == null)?"<null>":this.status));
        sb.append(',');
        sb.append("timeout");
        sb.append('=');
        sb.append(((this.timeout == null)?"<null>":this.timeout));
        sb.append(',');
        sb.append("createdAt");
        sb.append('=');
        sb.append(((this.createdAt == null)?"<null>":this.createdAt));
        sb.append(',');
        sb.append("updatedAt");
        sb.append('=');
        sb.append(((this.updatedAt == null)?"<null>":this.updatedAt));
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
        result = ((result* 31)+((this.satId == null)? 0 :this.satId.hashCode()));
        result = ((result* 31)+((this.url == null)? 0 :this.url.hashCode()));
        result = ((result* 31)+((this.playbookRunUrl == null)? 0 :this.playbookRunUrl.hashCode()));
        result = ((result* 31)+((this.timeout == null)? 0 :this.timeout.hashCode()));
        result = ((result* 31)+((this.labels == null)? 0 :this.labels.hashCode()));
        result = ((result* 31)+((this.playbookName == null)? 0 :this.playbookName.hashCode()));
        result = ((result* 31)+((this.createdAt == null)? 0 :this.createdAt.hashCode()));
        result = ((result* 31)+((this.service == null)? 0 :this.service.hashCode()));
        result = ((result* 31)+((this.recipient == null)? 0 :this.recipient.hashCode()));
        result = ((result* 31)+((this.correlationId == null)? 0 :this.correlationId.hashCode()));
        result = ((result* 31)+((this.id == null)? 0 :this.id.hashCode()));
        result = ((result* 31)+((this.additionalProperties == null)? 0 :this.additionalProperties.hashCode()));
        result = ((result* 31)+((this.satOrgId == null)? 0 :this.satOrgId.hashCode()));
        result = ((result* 31)+((this.account == null)? 0 :this.account.hashCode()));
        result = ((result* 31)+((this.status == null)? 0 :this.status.hashCode()));
        result = ((result* 31)+((this.updatedAt == null)? 0 :this.updatedAt.hashCode()));
        return result;
    }

    @Override
    public boolean equals(Object other) {
        if (other == this) {
            return true;
        }
        if ((other instanceof Payload) == false) {
            return false;
        }
        Payload rhs = ((Payload) other);
        return (((((((((((((((((this.satId == rhs.satId)||((this.satId!= null)&&this.satId.equals(rhs.satId)))&&((this.url == rhs.url)||((this.url!= null)&&this.url.equals(rhs.url))))&&((this.playbookRunUrl == rhs.playbookRunUrl)||((this.playbookRunUrl!= null)&&this.playbookRunUrl.equals(rhs.playbookRunUrl))))&&((this.timeout == rhs.timeout)||((this.timeout!= null)&&this.timeout.equals(rhs.timeout))))&&((this.labels == rhs.labels)||((this.labels!= null)&&this.labels.equals(rhs.labels))))&&((this.playbookName == rhs.playbookName)||((this.playbookName!= null)&&this.playbookName.equals(rhs.playbookName))))&&((this.createdAt == rhs.createdAt)||((this.createdAt!= null)&&this.createdAt.equals(rhs.createdAt))))&&((this.service == rhs.service)||((this.service!= null)&&this.service.equals(rhs.service))))&&((this.recipient == rhs.recipient)||((this.recipient!= null)&&this.recipient.equals(rhs.recipient))))&&((this.correlationId == rhs.correlationId)||((this.correlationId!= null)&&this.correlationId.equals(rhs.correlationId))))&&((this.id == rhs.id)||((this.id!= null)&&this.id.equals(rhs.id))))&&((this.additionalProperties == rhs.additionalProperties)||((this.additionalProperties!= null)&&this.additionalProperties.equals(rhs.additionalProperties))))&&((this.satOrgId == rhs.satOrgId)||((this.satOrgId!= null)&&this.satOrgId.equals(rhs.satOrgId))))&&((this.account == rhs.account)||((this.account!= null)&&this.account.equals(rhs.account))))&&((this.status == rhs.status)||((this.status!= null)&&this.status.equals(rhs.status))))&&((this.updatedAt == rhs.updatedAt)||((this.updatedAt!= null)&&this.updatedAt.equals(rhs.updatedAt))));
    }

    public enum Status {

        RUNNING("running"),
        SUCCESS("success"),
        FAILURE("failure"),
        TIMEOUT("timeout"),
        CANCELED("canceled");
        private final String value;
        private final static Map<String, Payload.Status> CONSTANTS = new HashMap<String, Payload.Status>();

        static {
            for (Payload.Status c: values()) {
                CONSTANTS.put(c.value, c);
            }
        }

        private Status(String value) {
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
        public static Payload.Status fromValue(String value) {
            Payload.Status constant = CONSTANTS.get(value);
            if (constant == null) {
                throw new IllegalArgumentException(value);
            } else {
                return constant;
            }
        }

    }

}
