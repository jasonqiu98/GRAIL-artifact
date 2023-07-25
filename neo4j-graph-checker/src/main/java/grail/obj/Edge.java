package grail.obj;

import com.fasterxml.jackson.annotation.JsonAnySetter;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;

public class Edge {
    @JsonProperty
    private String _from;
    @JsonProperty
    private String _to;
    @JsonIgnore
    private String _key;
    @JsonIgnore
    private String _id;
    @JsonIgnore
    private String _rev;
    @JsonProperty
    private String from_evt;
    @JsonProperty
    private String to_evt;
    private String type;

    public String getType() {
        return type;
    }

    public String get_from() {
        return _from;
    }

    public String get_to() {
        return _to;
    }

    public String getFrom_evt() {
        return from_evt;
    }

    public String getTo_evt() {
        return to_evt;
    }

    @JsonAnySetter
    public void set_key(String _key) {
        this._key = _key;
    }

    @JsonAnySetter
    public void set_id(String _id) {
        this._id = _id;
    }

    @JsonAnySetter
    public void set_rev(String _rev) {
        this._rev = _rev;
    }

    @JsonAnySetter
    public void setFrom_evt(String from_evt) {
        this.from_evt = from_evt;
    }

    @JsonAnySetter
    public void setTo_evt(String to_evt) {
        this.to_evt = to_evt;
    }

    @JsonAnySetter
    public void setType(String type) {
        this.type = type;
    }

    @JsonAnySetter
    public void set_from(String _from) {
        this._from = _from;
    }

    @JsonAnySetter
    public void set_to(String _to) {
        this._to = _to;
    }

    @Override
    public String toString() {
        return _from + "->" + _to;
    }
}
