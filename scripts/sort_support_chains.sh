#!/usr/bin/env bash

yq eval 'sort_keys(.)' -i ./docker/cvms/support_chains.yaml