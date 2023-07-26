(ns jepsen.arangodb.utils.support
  (:require [clojure.tools.logging :refer :all]
            [clojure.string :as str]
            [jepsen [control :as c]
             [db :as db]]
            [jepsen.control.util :as cu]
            [jepsen.tests :as tests]
            [jepsen.os.debian :as debian]))

(def dir "/opt/arangodb")
(def bin-dir (str dir "/bin"))
(def data-dir (str bin-dir "/data"))
(def binary "arangodb")
(def logfile "/home/vagrant/arangodb.log")
(def pidfile "/home/vagrant/arangodb.pid")

(def jwt-secret-path "/home/vagrant/arangodb.secret")

(defn cli-arg
  "command line argument of arangodb"
  [flag value]
  (str flag "=" value))

(defn parse-int [s]
  (Integer/parseInt (re-find #"\A-?\d+" s)))

(defn db-setup
  "ArangoDB Version v3.9.10"
  []
  (reify db/DB
    (setup! [_ test node]
      (info node "installing arangodb v3.9.10")
      (c/su
       (let [url (str "https://download.arangodb.com/arangodb39/Community/Linux/arangodb3-linux-3.9.10.tar.gz")]
         (cu/install-archive! url dir)))

      ; /opt/arangodb/bin/arangodb --starter.mode=single --server.storage-engine=rocksdb --auth.jwt-secret=/home/vagrant/arangodb.secret --starter.data-dir=./data
      (c/su
       (cu/start-daemon!
        {:logfile logfile
         :pidfile pidfile
         :chdir bin-dir}
        binary
        (cli-arg "--starter.mode" "single")
        (cli-arg "--server.storage-engine" "rocksdb")
        (cli-arg "--auth.jwt-secret" jwt-secret-path)
        (cli-arg "--starter.data-dir" data-dir))))

    (teardown! [_ test node]
      (info node "tearing down arangodb")
      (cu/stop-daemon! binary pidfile)
      (try
        ; can use cu/grepkill! instead
        ; https://github.com/jepsen-io/jepsen/blob/40b24800122433ea260bd188c05033059329d3a0/jepsen/src/jepsen/control/util.clj#L286
        (let [arango-procs (str/split (c/exec "pgrep" "arango") #"\n")]
          (c/su (c/exec "kill" "-9" arango-procs))
          (info "processes" arango-procs "killed")
          ; delete the data directory
          (c/su (c/exec :rm :-rf data-dir)))
        (catch clojure.lang.ExceptionInfo e
          (info "no arangodb processes killed"))))

    db/LogFiles
    (log-files [_ test node]
      [logfile])))

(def basic-test
  "Given an options map from the command line runner (e.g. :nodes, :ssh,
  :concurrency, ...), constructs a test map."
  (merge tests/noop-test
         {:name "arangodb-basic-test"
          :os debian/os
          :db (db-setup)
          :pure-generators true}))