package grail.obj;

import com.fasterxml.jackson.annotation.JsonAnySetter;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;

public class Vertex {
    @JsonProperty
    private String _id;
    @JsonIgnore
    private String _key;
    @JsonIgnore
    private String _rev;
    @JsonIgnore
    private int index;

    @JsonAnySetter
    public void set_id(String _id) {
        this._id = _id;
    }

    public String get_id() {
        return _id;
    }

    @JsonAnySetter
    public void set_key(String _key) {
        this._key = _key;
    }

    @JsonAnySetter
    public void set_rev(String _rev) {
        this._rev = _rev;
    }

    @JsonAnySetter
    public void setIndex(int index) {
        this.index = index;
    }

    @Override
    public String toString() {
        return _id;
    }
}
