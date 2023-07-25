package grail.graph;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;

import grail.obj.Edge;
import grail.obj.Vertex;

import java.io.IOException;
import java.util.List;

public class ImportUtil {
    private final static ObjectMapper om = new ObjectMapper();

    public static List<Vertex> getVertex(String vertex) throws IOException {
        return om.readValue(ImportUtil.class.getClassLoader().getResourceAsStream(vertex), new TypeReference<List<Vertex>>(){});
    }

    public static List<Edge> getEdge(String edge) throws IOException {
        return om.readValue(ImportUtil.class.getClassLoader().getResourceAsStream(edge), new TypeReference<List<Edge>>() {});
    }
}
