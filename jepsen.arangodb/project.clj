(defproject jepsen.arangodb "0.1.0-SNAPSHOT"
  :description "A Jepsen test for ArangoDB"
  :url "https://github.com/grail/jepsen.arangodb.git"
  :license {:name "EPL-2.0 OR GPL-2.0-or-later WITH Classpath-exception-2.0"
            :url "https://www.eclipse.org/legal/epl-2.0/"}
  :main jepsen.arangodb
  :dependencies [[org.clojure/clojure "1.11.1"]
                 [jepsen "0.2.7"]
                 [com.arangodb/arangodb-java-driver "6.16.0"]
                 [org.apache.httpcomponents/httpclient "4.5.13"]]
  :repl-options {:init-ns jepsen.arangodb})
