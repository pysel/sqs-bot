import time
import pytest

import conftest
from sqs_service import *
from quote import *
from quote_response import *
from rand_util import *
from e2e_math import *
from decimal import *
from constants import *
from util import *
from route import *

# Arbitrary choice based on performance at the time of test writing
EXPECTED_LATENCY_UPPER_BOUND_MS = 15000

# Test suite for the /router/custom-direct-quote endpoint
# Test runs tests for exact amount out quotes.
class TestExactAmountOutDirectCustomQuote:
    @pytest.mark.parametrize("pair", conftest.create_coins_from_pairs(conftest.create_no_dupl_token_pairs(conftest.choose_tokens_liq_range(num_tokens=10, min_liq=500_000, exponent_filter=USDC_PRECISION)), USDC_PRECISION, USDC_PRECISION + 3), ids=id_from_swap_pair)
    def test_get_custom_direct_quote(self, environment_url, pair):
        TestExactAmountOutDirectCustomQuote.run_get_custom_direct_quote(environment_url, pair['token_in']['amount_str'], pair['token_in']['denom'], pair['out_denom'])

    @staticmethod
    def run_get_custom_direct_quote(environment_url, amount_str, token_out_denom, denom_in):
        coin = Coin(token_out_denom, amount_str)
        token_out  = amount_str + token_out_denom

        # Get the optimal quote for the given token pair
        # Direct custom quote does not support multiple routes, so we request single/multi hop pool routes only
        optimal_quote = ExactAmountOutQuote.run_quote_test(environment_url, token_out, denom_in, False, True, EXPECTED_LATENCY_UPPER_BOUND_MS)

        pool_id = ','.join(map(str, optimal_quote.get_pool_ids()))
        denoms_in = ','.join(map(str, optimal_quote.get_token_in_denoms()))

        quote = TestExactAmountOutDirectCustomQuote.run_quote_test(environment_url, token_out, denoms_in, pool_id, EXPECTED_LATENCY_UPPER_BOUND_MS)

        # All tokens have the same default exponent, resulting in scaling factor of 1.
        spot_price_scaling_factor = 1

        expected_in_base_out_quote_price, expected_token_in, token_out_amount_usdc_value = ExactAmountOutQuote.calculate_expected_base_out_quote_spot_price(denom_in, coin)

        # Chose the error tolerance based on amount in swapped.
        error_tolerance = Quote.choose_error_tolerance(token_out_amount_usdc_value)

        # Validate that price impact is present.
        assert quote.price_impact is not None

        # Validate quote results
        ExactAmountOutQuote.validate_quote_test(quote, coin.amount, coin.denom, spot_price_scaling_factor, expected_in_base_out_quote_price, expected_token_in, denom_in, error_tolerance, True)

    @staticmethod
    def run_quote_test(environment_url, token_out, denom_in, pool_id, expected_latency_upper_bound_ms, expected_status_code=200) -> QuoteExactAmountOutResponse:
        """
        Runs exact amount out test for the /router/custom-direct-quote endpoint with the given input parameters.
        """

        service_call = lambda: conftest.SERVICE_MAP[environment_url].get_exact_amount_out_custom_direct_quote(token_out, denom_in, pool_id)

        response = Quote.run_quote_test(service_call, expected_latency_upper_bound_ms, expected_status_code)

        # Return route for more detailed validation
        return QuoteExactAmountOutResponse(**response)

