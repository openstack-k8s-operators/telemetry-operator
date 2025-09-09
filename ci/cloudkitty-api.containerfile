FROM quay.rdoproject.org/podified-master-centos10/openstack-cloudkitty-api:current-tested

ARG CK_REPO="git+https://review.opendev.org/openstack/cloudkitty"
ARG CK_BRANCH=master

RUN python3 -m ensurepip && python3 -m pip install --prefix=/usr $CK_REPO@$CK_BRANCH

