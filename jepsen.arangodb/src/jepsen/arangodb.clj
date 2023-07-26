(ns jepsen.arangodb
  (:require [jepsen.arangodb.tests [register :as register]
             [list-append :as la]]
            [jepsen [cli :as cli]]
            [jepsen.tests :as tests]))

(def cli-opts
  "Additional command line options."
  [["-r" "--rate HZ" "Approximate number of requests per second, per thread."
    :default  1
    :parse-fn read-string
    :validate [#(and (number? %) (pos? %)) "Must be a positive number"]]
   ; register-test independent-gen related
   [nil "--ops-per-key NUM" "Maximum number of operations on any given key."
    :default  20
    :parse-fn parse-long
    :validate [pos? "Must be a positive integer."]]
   [nil "--threads-per-group NUM" "Number of threads, per group."
    :default  5
    :parse-fn parse-long
    :validate [pos? "Must be a positive integer."]]
   ; list-append-test related
   [nil "--key-count NUM" "Number of distinct keys at any point."
    :default  5
    :parse-fn parse-long
    :validate [pos? "Must be a positive integer."]]
   [nil "--min-txn-length NUM" "Minimum number of operations per txn."
    :default  4
    :parse-fn parse-long
    :validate [pos? "Must be a positive integer."]]
   [nil "--max-txn-length NUM" "Maximum number of operations per txn."
    :default  8
    :parse-fn parse-long
    :validate [pos? "Must be a positive integer."]]
   [nil "--max-writes-per-key NUM" "Maximum number of operations per key."
    :default  3
    :parse-fn parse-long
    :validate [pos? "Must be a positive integer."]]
   [nil "--nemesis-type partition|noop" "Nemesis used."
    :default :noop
    :parse-fn #(case %
                 ("partition") :partition
                 ("noop") :noop
                 :invalid)
    :validate [#{:partition :noop} "Unsupported nemesis"]]
   [nil "--test-type register|la" "Test type used."
    :default :invalid
    :parse-fn #(case %
                 ("register") :register
                 ("la") :list-append
                 :invalid)
    :validate [#{:register :list-append} "Unsupported test type"]]])

(defn single-test-wrapper
  "a wrapper from the register test and list append test"
  [opts]
  (let [test-fn (case (:test-type opts)
                  :register register/register-test
                  :list-append la/list-append-test
                  tests/noop-test)]
    (test-fn opts)))

(defn -main
  "Handles command line arguments. Can either run a test, or a web server for browsing results"
  [& args]
  (cli/run! (merge (cli/single-test-cmd {:test-fn single-test-wrapper
                                         :opt-spec cli-opts})
                   (cli/serve-cmd))
            args))
