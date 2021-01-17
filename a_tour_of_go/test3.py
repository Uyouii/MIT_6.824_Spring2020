#coding: utf8

import math
import scipy.stats 

def EloTest(Ra, Rb):
    Ea = 1 / (1 + math.pow(10, (Rb - Ra)/400.0))
    Eb = 1 / (1 + math.pow(10, (Ra - Rb)/400.0))

    k = 30

    cRa = 30 * (1 - Ea)
    cRb = 30 * ( 0 - Eb)

    print Ea, Eb

def TrueSkillTest(mu1, mu2, sigma1, sigma2):
    sigma0 = 1000 / 3.0
    mu0 = 1000
    beta = sigma0 / 2.0
    beta_2 = math.pow(beta, 2)

    c = 2 * beta_2 + math.pow(sigma1, 2) + math.pow(sigma2, 2)

    draw = math.pow(2 * beta_2 / c, 0.5) * math.exp(-1 * math.pow(mu1 - mu2, 2) / (2 * c))

    print draw

    print scipy.stats.norm.cdf((mu1 - mu2) / math.sqrt(c))


if __name__ == "__main__":
    TrueSkillTest(1000.0, 1500.0, 200, 300)
