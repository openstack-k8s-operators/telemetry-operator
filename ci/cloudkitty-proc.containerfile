FROM quay.rdoproject.org/podified-master-centos10/openstack-cloudkitty-processor:current-tested

ARG CK_REPO="git+https://review.opendev.org/openstack/cloudkitty"
ARG CK_BRANCH=master

USER root
RUN rm -rf /usr/lib/python3.12/site-packages/cloudkitty
RUN python3 -m ensurepip && python3 -m pip install --prefix=/usr $CK_REPO@$CK_BRANCH
USER cloudkitty
