(ns jepsen.arangodb.tests.list-append
  "Detects cycles in histories where operations are transactions
   over named lists, and operations are either appends or reads.
   See elle.list-append for docs."
  (:require [clojure.tools.logging :refer :all]
            [jepsen [checker :as checker]
             [client :as client]
             [generator :as gen]
             [nemesis :as nemesis]]
            [jepsen.arangodb.utils [driver :as driver]
             [support :as s]]
            [jepsen.checker.timeline :as timeline]
            [jepsen.tests.cycle.append :as la])
  (:import com.arangodb.model.StreamTransactionOptions))

(def dbName "listAppend")
(def collectionName "laCol")
(def attributeName "laAttr")

(defrecord Client [db-created? collection-created? conn node]
  client/Client
  (open! [this test node]
    (assoc this :conn (-> (new com.arangodb.ArangoDB$Builder)
                          (.host node 8529)
                          (.user "root")
                          (.password "")
                          (.timeout (int 10000)) ; 10s timeout for connection and request
                          (.build))
           :node (str node)))

  (setup! [this test]
    (info "sleep 15s to make sure the connection is ready")
    ; you may need to adjust the duration for system discrepancies 
    (Thread/sleep 15000)
    (locking db-created?
      (info "Prepare to create databases")
      (while (false? (compare-and-set! db-created? true false))
        (info "Creating databases")
        (try (Thread/sleep 500)
             ;; create a database with the name of `dbName`
             (-> conn (driver/create-db dbName))
             (info "Databases created")
             (reset! db-created? true)
             ;; database not created yet
             (catch java.lang.NullPointerException e
               (warn "Databases not created yet")
               (Thread/sleep 2000))
             ;; database already created
             (catch com.arangodb.ArangoDBException ex
               (warn (.getErrorMessage ex))
               (reset! db-created? true)))))

    (locking collection-created?
      (while (false? (compare-and-set! collection-created? true false))
        (info "Creating collections")
        (try (Thread/sleep 500)
             (-> conn (driver/create-collection dbName collectionName))
             (info "Collections created")
             (reset! collection-created? true)
             (catch java.lang.NullPointerException e
               (warn "Collections not created yet")
               (Thread/sleep 2000))
             (catch com.arangodb.ArangoDBException ex
               (warn (.getErrorMessage ex))
               (reset! collection-created? true))))))

  (invoke! [this test op]
    ; operation from {:type :invoke, :f :txn, :value [[:r 3 nil] [:append 3 2] [:r 3]]}
    ; to {:type :ok, :f :txn, :value [[:r 3 [1]] [:append 3 2] [:r 3 [1 2]]]}
    (let [txn-vec (:value op)
          db (-> conn (driver/get-db dbName))
          txn-entity (.beginStreamTransaction db (.writeCollections (new StreamTransactionOptions) (into-array [collectionName])))
          txn-id (.getId txn-entity)]
      (try
        (let [ret-val (driver/submit-txn-la conn dbName collectionName attributeName txn-vec txn-entity)]
          (if (not= ret-val nil)
            (do (.commitStreamTransaction db txn-id)
                (info (str "committed transaction " txn-id))
                (assoc op :type :ok :value ret-val))
            (do (.abortStreamTransaction db txn-id)
                (info (str "aborted transaction " txn-id))
                (assoc op :type :fail))))
        (catch java.net.SocketTimeoutException ex
          (.abortStreamTransaction db txn-id)
          (info (str "aborted transaction " txn-id))
          (assoc op :type :fail, :error :timeout))
        (catch java.lang.NullPointerException ex
          (.abortStreamTransaction db txn-id)
          (info (str "aborted transaction " txn-id))
          (error "Connection error")
          (assoc op
                 :type  (if (= :read (:f op)) :fail :info)
                 :error :connection-lost))
        (catch com.arangodb.ArangoDBException ex
          (.abortStreamTransaction db txn-id)
          (info (str "aborted transaction " txn-id))
          (warn (.getErrorMessage ex))
        ;; 1200 write-write conflict; key: Anna
        ;; 1465 cluster internal HTTP connection broken
        ;; 1457 timeout in cluster operation
          (let [errorCodeMap
                {1200 :ww-conflict
                 1465 :conn-closed
                 1457 :timeout
                 nil  :timeout}
                errorCode (.getErrorNum ex)]
            (assoc op :type :fail, :error (get errorCodeMap errorCode errorCode)))))))

  (teardown! [this test]
    (try
      (.shutdown conn)
      (info "Connection closed")
      (catch clojure.lang.ExceptionInfo e
        (info "Connection not closed!"))))

  (close! [_ test]))

(defn list-append-test
  "Given an options map from the command line runner (e.g. :nodes, :ssh,
  :concurrency, ...), constructs a test map."
  [opts]
  (merge s/basic-test
         opts
         {:name            "arangodb-list-append-test"
          :client          (Client. (atom false) (atom false) nil nil)
          :nemesis         (case (:nemesis-type opts)
                             :partition (nemesis/partition-random-halves)
                             :noop nemesis/noop)
          ;; from https://github.com/jepsen-io/jepsen/blob/main/jepsen/src/jepsen/tests/cycle/append.clj
          :generator       (let [la-gen (->> (la/gen opts)
                                             (gen/stagger (/ (:rate opts))))]
                             (case (:nemesis-type opts)
                               :partition (->> la-gen
                                               (gen/nemesis
                                                (cycle [(gen/sleep 5)
                                                        {:type :info, :f :start}
                                                        (gen/sleep 5)
                                                        {:type :info, :f :stop}]))
                                               (gen/time-limit (:time-limit opts)))
                               :noop (->> la-gen
                                          (gen/nemesis nil)
                                          (gen/time-limit (:time-limit opts)))))
          :checker         (checker/compose
                            {:la     (la/checker)
                             :perf   (checker/perf)
                             :timeline (timeline/html)})}))