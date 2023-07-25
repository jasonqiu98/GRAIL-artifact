package grail;

import org.neo4j.driver.Session;

public class ImportRunner {
    public static String RW = "MATCH ()-[r:rw]->() RETURN count(r)";
    public static String WW = "MATCH ()-[r:ww]->() RETURN count(r)";
    public static String WR = "MATCH ()-[r:wr]->() RETURN count(r)";
    public static String ALL = "MATCH ()-[r]->() RETURN count(r)";
    public static String NODES = "MATCH (n:txn) RETURN count(n)";
    public static void main(String[] args) throws Exception{
        String[] list = new String[]{"rw_register"};
        for (String dir : list) {
            System.out.println(dir);
            for (int i = 10; i <= 200; i+=10) {
                Application.importGraph(dir, i);
                try (Session session = Application.driver.session()){
                    int nodes = session.run(NODES).next().get(0).asInt();
                    int edges = session.run(ALL).next().get(0).asInt();
                    int rws = session.run(RW).next().get(0).asInt();
                    int wws = session.run(WW).next().get(0).asInt();
                    int wrs = session.run(WR).next().get(0).asInt();
                    System.out.println(nodes+"\t"+edges+"\t"+rws+"\t"+wws+"\t"+wrs);
                }
            }
        }
    }
}
