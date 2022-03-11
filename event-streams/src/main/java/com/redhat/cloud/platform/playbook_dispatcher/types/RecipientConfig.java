
package com.redhat.cloud.platform.playbook_dispatcher.types;

import java.util.HashMap;
import java.util.Map;
import com.fasterxml.jackson.annotation.JsonAnyGetter;
import com.fasterxml.jackson.annotation.JsonAnySetter;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonPropertyOrder;

@JsonInclude(JsonInclude.Include.NON_NULL)
@JsonPropertyOrder({
    "sat_id",
    "sat_org_id"
})
public class RecipientConfig {

    @JsonProperty("sat_id")
    private String satId;
    @JsonProperty("sat_org_id")
    private String satOrgId;
    @JsonIgnore
    private Map<String, Object> additionalProperties = new HashMap<String, Object>();

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
        sb.append(RecipientConfig.class.getName()).append('@').append(Integer.toHexString(System.identityHashCode(this))).append('[');
        sb.append("satId");
        sb.append('=');
        sb.append(((this.satId == null)?"<null>":this.satId));
        sb.append(',');
        sb.append("satOrgId");
        sb.append('=');
        sb.append(((this.satOrgId == null)?"<null>":this.satOrgId));
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
        result = ((result* 31)+((this.additionalProperties == null)? 0 :this.additionalProperties.hashCode()));
        result = ((result* 31)+((this.satOrgId == null)? 0 :this.satOrgId.hashCode()));
        return result;
    }

    @Override
    public boolean equals(Object other) {
        if (other == this) {
            return true;
        }
        if ((other instanceof RecipientConfig) == false) {
            return false;
        }
        RecipientConfig rhs = ((RecipientConfig) other);
        return ((((this.satId == rhs.satId)||((this.satId!= null)&&this.satId.equals(rhs.satId)))&&((this.additionalProperties == rhs.additionalProperties)||((this.additionalProperties!= null)&&this.additionalProperties.equals(rhs.additionalProperties))))&&((this.satOrgId == rhs.satOrgId)||((this.satOrgId!= null)&&this.satOrgId.equals(rhs.satOrgId))));
    }

}
