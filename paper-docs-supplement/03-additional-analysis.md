# Additional Scalability Analysis

Following the experiments related to collection time and rate, we are also able to extend the scalability analysis and analyze the effects of three underlying factors (history length, density, number of contributing traversals) on the runtime of the checker. We combine the execution histories across multiple datasets with varying history length and rate in both list and register cases. These datasets include DS1 and DS3 in list cases, and two additional datasets in register cases. Also, we record the values of the three underlying factors for each history. After that, we perform linear regression between each underlying factor and the checker's runtime to observe the correlation in a standard linear model. In the following table, we show that all the three underlying factor are significant in serializability checking across the four datasets, with all the three p-values below the critical cut-off of 0.05. We have also extended the analysis to SI checking and found the significance of the three factors.

| $T_{ser}$ | coefficients | p-value                 |
|-----------|--------------|-------------------------|
| Intercept | $-588.42$    | $0.078$                 |
| $L$       | $0.27$       | $3.50 \times 10^{-41}$  |
| $D$       | $186.41$     | $0.049$                 |
| $N$       | $-0.64$      | $2.62 \times 10^{-16}$  |

